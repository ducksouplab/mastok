package live

import (
	"log"
	"strconv"

	"github.com/ducksouplab/mastok/models"
	"github.com/ducksouplab/mastok/otree"
)

// clients hold references to any (supervisor or participant) client
// the currentPool holds count of participant clients
type runner struct {
	campaign         *models.Campaign
	poolSize         int
	clients          map[*client]bool
	poolFingerprints map[string]bool
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
		poolSize:         0,
		clients:          make(map[*client]bool),
		poolFingerprints: map[string]bool{}, // used only for JoinOnce campaigns
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

func (r *runner) stateMessage(c *client) Message {
	return Message{
		Kind:    "State",
		Payload: r.campaign.GetPublicState(c.isSupervisor),
	}
}

func (r *runner) poolSizeMessage() Message {
	return Message{
		Kind:    "PoolSize",
		Payload: strconv.Itoa(r.poolSize) + "/" + strconv.Itoa(r.campaign.PerSession),
	}
}

func sessionStartParticipantMessage(code string) Message {
	return Message{
		Kind:    "SessionStart",
		Payload: otree.ParticipantStartURL(code),
	}
}

func sessionStartSupervisorMessage(session models.Session) Message {
	return Message{
		Kind:    "SessionStart",
		Payload: session,
	}
}

func participantDisconnectMessage() Message {
	return Message{
		Kind: "Disconnect",
	}
}

func participantRejectMessage() Message {
	return Message{
		Kind: "Reject",
	}
}

func participantRedirectMessage(code string) Message {
	return Message{
		Kind:    "Redirect",
		Payload: otree.ParticipantStartURL(code),
	}
}

func (r *runner) tickStateMessage() {
	if r.updateStateTicker != nil {
		r.updateStateTicker.stop()
	}
	ticker := newTicker(models.SessionDurationUnit)
	go ticker.loop(r)
	r.updateStateTicker = ticker
}

func (r *runner) loop() {
	for {
		select {
		case c := <-r.registerCh:
			if r.campaign.State == "Running" || c.isSupervisor {
				r.clients[c] = true
				c.outgoingCh <- r.stateMessage(c)

				if c.isSupervisor {
					// only inform supervisor client about the pool size right away
					c.outgoingCh <- r.poolSizeMessage()
				}
			} else {
				// don't register
				c.outgoingCh <- r.stateMessage(c)
				c.outgoingCh <- participantDisconnectMessage()
				if len(r.clients) == 0 {
					r.stop()
					return
				}
			}
		case c := <-r.unregisterCh:
			if _, ok := r.clients[c]; ok {
				// leaves pool if participant has joined
				if !c.isSupervisor && c.hasJoinedPool {
					r.poolSize -= 1
					// tells everyone including supervisor
					for c := range r.clients {
						c.outgoingCh <- r.poolSizeMessage()
					}
				}
				// actually deletes client
				delete(r.clients, c)
				if c.runner.campaign.JoinOnce {
					delete(r.poolFingerprints, c.fingerprint)
				}
				if len(r.clients) == 0 {
					r.stop()
					return
				}
			}
		case c := <-r.landCh:
			if c.runner.campaign.JoinOnce {
				if _, ok := r.poolFingerprints[c.fingerprint]; ok { // is in pool?
					c.outgoingCh <- participantRejectMessage()
				} else {
					// process if reply is needed
					// p, err := models.FindParticipation(*c.runner.campaign, c.fingerprint)
					// log.Printf(">>>>>> p %#v %#v", p, err)
					// if err == nil {
					// 	s, err := models.FindSession(p.SessionID)
					// 	if err == nil {
					// 		log.Printf(">>>>>> s %#v %#v %#v", s, s.IsLive(), err)
					// 		if s.IsLive() {
					// 			log.Printf(">>>>>> live")
					// 			c.outgoingCh <- participantRedirectMessage(p.OtreeCode)
					// 		} else {
					// 			c.outgoingCh <- participantRejectMessage()
					// 		}
					// 	}
					// }
					r.poolFingerprints[c.fingerprint] = true
				}
			}
		case m := <-r.incomingCh:
			if m.Kind == "State" {
				r.campaign.State = m.Payload.(string)
				models.DB.Save(r.campaign)
				for c := range r.clients {
					newMessageState := r.stateMessage(c)
					c.outgoingCh <- newMessageState
					if newMessageState.Payload == "Unavailable" {
						c.outgoingCh <- participantDisconnectMessage()
					}
				}
			} else if m.Kind == "Join" {
				// it's a partcipant -> increases pool
				r.poolSize += 1
				// inform everyone (participants and supervisors) about the new pool size
				for c := range r.clients {
					c.outgoingCh <- r.poolSizeMessage()
				}
				// starts session when pool is full
				if r.poolSize == r.campaign.PerSession {
					session, participantCodes, err := models.CreateSession(r.campaign)
					if err != nil {
						log.Println("[runner] session creation failed: ", err)
					} else {
						participantIndex := 0
						for c := range r.clients {
							if c.isSupervisor {
								c.outgoingCh <- sessionStartSupervisorMessage(session)
								c.outgoingCh <- r.stateMessage(c) // if state becomes Busy
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
