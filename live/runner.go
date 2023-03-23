package live

import (
	"log"

	"github.com/ducksouplab/mastok/models"
)

// clients hold references to any (supervisor or participant) client
type runner struct {
	campaign         *models.Campaign
	roomSize         int
	clients          map[*client]bool
	roomFingerprints map[string]bool
	// manage broadcasting
	registerCh   chan *client
	unregisterCh chan *client
	// other incoming events
	landCh     chan *client
	incomingCh chan Message
	// done
	updateStateTicker *ticker
	doneCh            chan struct{}
}

func newRunner(c *models.Campaign) *runner {
	r := runner{
		campaign:         c,
		roomSize:         0,
		clients:          make(map[*client]bool),
		roomFingerprints: map[string]bool{}, // used only for JoinOnce campaigns
		registerCh:       make(chan *client),
		unregisterCh:     make(chan *client),
		landCh:           make(chan *client),
		incomingCh:       make(chan Message),
		doneCh:           make(chan struct{}),
	}

	if c.GetPublicState(true) == models.Busy {
		r.tickStateMessage()
	}

	return &r
}

func (r *runner) isDone() chan struct{} {
	return r.doneCh
}

func (r *runner) stop() {
	deleteRunner(r.campaign)
	close(r.doneCh)
}

func (r *runner) loop() {
	for {
		select {
		case c := <-r.registerCh:
			if r.campaign.State == "Running" || c.isSupervisor {
				r.clients[c] = true
				c.outgoingCh <- stateMessage(r.campaign, c)

				if c.isSupervisor {
					// only inform supervisor client about the room size right away
					c.outgoingCh <- roomSizeMessage(r)
				}
			} else {
				// don't register
				c.outgoingCh <- stateMessage(r.campaign, c)
				c.outgoingCh <- participantDisconnectMessage()
				if len(r.clients) == 0 {
					r.stop()
					return
				}
			}
		case c := <-r.unregisterCh:
			if _, ok := r.clients[c]; ok {
				// leaves room if participant has joined
				if !c.isSupervisor && c.hasJoinedRoom {
					r.roomSize -= 1
					// tells everyone including supervisor
					for c := range r.clients {
						c.outgoingCh <- roomSizeMessage(r)
					}
				}
				// actually deletes client
				delete(r.clients, c)
				if c.runner.campaign.JoinOnce {
					delete(r.roomFingerprints, c.fingerprint)
				}
				if len(r.clients) == 0 {
					r.stop()
					return
				}
			}
		case c := <-r.landCh:
			var isInLiveSession bool
			participation, hasParticipatedToCampaign := models.GetParticipation(*c.runner.campaign, c.fingerprint)
			if hasParticipatedToCampaign {
				pastSession, ok := models.GetSession(participation.SessionID)
				if ok {
					isInLiveSession = pastSession.IsLive()
				}
			}
			if isInLiveSession { // we assume it's a reconnect, so we redirect to oTree
				c.outgoingCh <- participantRedirectMessage(participation.OtreeCode)
				break
			}
			if c.runner.campaign.JoinOnce {
				if _, ok := r.roomFingerprints[c.fingerprint]; ok { // is in room?
					c.outgoingCh <- participantRejectMessage()
					break
				}
				if hasParticipatedToCampaign { // has been found in one session of this campaign
					c.outgoingCh <- participantRejectMessage()
					break
				}
				r.roomFingerprints[c.fingerprint] = true
			}
			// finally lands in room
			c.outgoingCh <- participantConsentMessage(r.campaign)
		case m := <-r.incomingCh:
			if m.Kind == "State" {
				r.campaign.State = m.Payload.(string)
				models.DB.Save(r.campaign)
				for c := range r.clients {
					newMessageState := stateMessage(r.campaign, c)
					c.outgoingCh <- newMessageState
					if newMessageState.Payload == "Unavailable" {
						c.outgoingCh <- participantDisconnectMessage()
					}
				}
			} else if m.Kind == "Join" {
				// it's a partcipant -> increases room
				r.roomSize += 1
				// inform everyone (participants and supervisors) about the new room size
				for c := range r.clients {
					c.outgoingCh <- roomSizeMessage(r)
				}
				// starts session when room is full
				if r.roomSize == r.campaign.PerSession {
					session, participantCodes, err := models.CreateSession(r.campaign)
					if err != nil {
						log.Println("[runner] session creation failed: ", err)
					} else {
						participantIndex := 0
						for c := range r.clients {
							if c.isSupervisor {
								c.outgoingCh <- sessionStartSupervisorMessage(session)
								c.outgoingCh <- stateMessage(r.campaign, c) // if state becomes Busy
								if c.runner.campaign.GetPublicState(true) == models.Busy {
									c.runner.tickStateMessage()
								}
							} else {
								code := participantCodes[participantIndex]
								c.outgoingCh <- sessionStartParticipantMessage(code)
								models.CreateParticipation(session, c.fingerprint, code)
								c.outgoingCh <- participantDisconnectMessage()
								participantIndex++
							}
						}
					}
				}
			}
		}
	}
}
