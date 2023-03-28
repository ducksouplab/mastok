package live

import (
	"log"

	"github.com/ducksouplab/mastok/models"
)

//const maxWaitingPoolSize = 50

// clients hold references to any (supervisor or participant) client
type runner struct {
	campaign         *models.Campaign
	grouping         *models.Grouping // ready only, cached from campaign
	clients          *runnerClients
	roomFingerprints map[string]bool
	// manage broadcasting
	registerCh   chan *client
	unregisterCh chan *client
	// other incoming events
	participantCh chan FromParticipantMessage
	supervisorCh  chan Message
	// done
	updateStateTicker *ticker
	doneCh            chan struct{}
}

func newRunner(c *models.Campaign) *runner {
	g := c.GetGrouping()

	r := runner{
		campaign:         c,
		grouping:         g,
		clients:          newRunnerClients(g, c.PerSession),
		roomFingerprints: map[string]bool{}, // used only for JoinOnce campaigns
		registerCh:       make(chan *client),
		unregisterCh:     make(chan *client),
		// messages coming from participants or supervisor
		participantCh: make(chan FromParticipantMessage),
		supervisorCh:  make(chan Message),
		// signals end
		doneCh: make(chan struct{}),
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
				r.clients.add(c)
				c.outgoingCh <- stateMessage(r.campaign, c)

				if c.isSupervisor {
					// only inform supervisor client about the room size right away
					c.outgoingCh <- poolSizeMessage(r)
				}
			} else {
				// don't register
				c.outgoingCh <- stateMessage(r.campaign, c)
				c.outgoingCh <- participantDisconnectMessage()
				if r.clients.isEmpty() {
					r.stop()
					return
				}
			}
		case c := <-r.unregisterCh:
			if r.clients.has(c) {
				// deletes client
				if wasAgreeing := r.clients.delete(c); wasAgreeing {
					// tells everyone including supervisor that the room size has changed
					for c := range r.clients.pool {
						c.outgoingCh <- poolSizeMessage(r)
					}
					for c := range r.clients.supervisors {
						c.outgoingCh <- poolSizeMessage(r)
					}
				}
				if c.runner.campaign.JoinOnce {
					delete(r.roomFingerprints, c.fingerprint)
				}
				if r.clients.isEmpty() {
					r.stop()
					return
				}
			}
		case m := <-r.participantCh:
			if m.Kind == "Land" {
				var isInLiveSession bool
				participation, hasParticipated := models.GetParticipation(*r.campaign, m.Payload)
				if hasParticipated {
					pastSession, ok := models.GetSession(participation.SessionID)
					if ok {
						isInLiveSession = pastSession.IsLive()
					}
				}
				if isInLiveSession { // we assume it's a reconnect, so we redirect to oTree
					m.From.outgoingCh <- participantRedirectMessage(participation.OtreeCode)
					break
				}
				if r.campaign.JoinOnce {
					if _, ok := r.roomFingerprints[m.Payload]; ok { // is in room?
						m.From.outgoingCh <- participantRejectMessage()
						break
					}
					if hasParticipated { // has been found in one session of this campaign
						m.From.outgoingCh <- participantRejectMessage()
						break
					}
					r.roomFingerprints[m.Payload] = true
				}
				// finally lands in room
				m.From.outgoingCh <- participantConsentMessage(r.campaign)
			} else if m.Kind == "Choose" {
				r.clients.choose(m.From, m.Payload)
				// inform everyone (participants and supervisors) about the new room size
				for c := range r.clients.pool {
					c.outgoingCh <- poolSizeMessage(r)
				}
				for c := range r.clients.supervisors {
					c.outgoingCh <- poolSizeMessage(r)
				}
				// starts session when there is valid pool pending
				pool, ready := r.clients.tentativePool()
				if ready {
					session, participantCodes, err := models.CreateSession(r.campaign)
					if err != nil {
						log.Println("[runner] session creation failed: ", err)
					} else {
						participantIndex := 0
						for _, c := range pool {
							code := participantCodes[participantIndex]
							c.outgoingCh <- sessionStartParticipantMessage(code)
							models.CreateParticipation(session, c.fingerprint, code)
							c.outgoingCh <- participantDisconnectMessage()
							participantIndex++
						}
						for c := range r.clients.supervisors {
							c.outgoingCh <- sessionStartSupervisorMessage(session)
							c.outgoingCh <- stateMessage(r.campaign, c) // if state becomes Busy
							if c.runner.campaign.GetPublicState(true) == models.Busy {
								c.runner.tickStateMessage()
							}
						}
					}
				}
			}
		case m := <-r.supervisorCh:
			if m.Kind == "State" {
				r.campaign.State = m.Payload.(string)
				models.DB.Save(r.campaign)
				for c := range r.clients.all {
					newMessageState := stateMessage(r.campaign, c)
					c.outgoingCh <- newMessageState
					if newMessageState.Payload == "Unavailable" {
						c.outgoingCh <- participantDisconnectMessage()
					}
				}
			}
		}
	}
}
