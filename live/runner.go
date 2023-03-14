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
	campaign *models.Campaign
	poolSize int
	clients  map[*client]bool
	// manage broadcasting
	registerCh   chan *client
	unregisterCh chan *client
	// other incoming events
	stateCh chan string
	// done
	doneCh chan struct{}
}

func newRunner(c *models.Campaign) *runner {
	return &runner{
		campaign:     c,
		poolSize:     0,
		clients:      make(map[*client]bool),
		registerCh:   make(chan *client),
		unregisterCh: make(chan *client),
		stateCh:      make(chan string),
		doneCh:       make(chan struct{}),
	}
}

func (r *runner) isDone() chan struct{} {
	return r.doneCh
}

func (r *runner) stop() {
	deleteRunner(r.campaign)
	close(r.doneCh)
}

func (r *runner) stateMessage(c *client) Message {
	state := r.campaign.State
	// hide internals to participants
	if state != "Running" && !c.isSupervisor {
		state = "Unavailable"
	}
	return Message{
		Kind:    "State",
		Payload: state,
	}
}

func (r *runner) poolSizeMessage() Message {
	return Message{
		Kind:    "PoolSize",
		Payload: strconv.Itoa(r.poolSize) + "/" + strconv.Itoa(r.campaign.PerSession),
	}
}

func (r *runner) sessionStartParticipantMessage(code string) Message {
	return Message{
		Kind:    "SessionStart",
		Payload: otree.ParticipantStartURL(code),
	}
}

func (r *runner) sessionStartSupervisorMessage(session models.Session) Message {
	return Message{
		Kind:    "SessionStart",
		Payload: session,
	}
}

func participantDisconnectMessage() Message {
	return Message{
		Kind:    "Participant",
		Payload: "Disconnect",
	}
}

func (r *runner) loop() {
	for {
		select {
		case c := <-r.registerCh:
			if r.campaign.State == "Running" || c.isSupervisor {
				r.clients[c] = true
				c.messageCh <- r.stateMessage(c)

				if c.isSupervisor {
					// only inform supervisor client about the pool size
					c.messageCh <- r.poolSizeMessage()
				} else {
					// it's a partcipant -> increases pool
					r.poolSize += 1
					// inform everyone (participants and supervisors) about the new pool size
					for c := range r.clients {
						c.messageCh <- r.poolSizeMessage()
					}
					// starts session when pool is full
					if r.poolSize == r.campaign.PerSession {
						session, participantCodes, err := models.CreateSession(r.campaign)
						if err != nil {
							log.Println("[runner] oTree session creation failed")
						} else {
							participantIndex := 0
							for c := range r.clients {
								if c.isSupervisor {
									c.messageCh <- r.sessionStartSupervisorMessage(session)
								} else {
									c.messageCh <- r.sessionStartParticipantMessage(participantCodes[participantIndex])
									c.messageCh <- participantDisconnectMessage()
									participantIndex++
								}
							}
						}
					}
				}
			} else {
				c.messageCh <- r.stateMessage(c)
				c.messageCh <- participantDisconnectMessage()
				if len(r.clients) == 0 {
					r.stop()
					return
				}
			}
		case c := <-r.unregisterCh:
			if _, ok := r.clients[c]; ok {
				// leaves pool if participant
				if !c.isSupervisor {
					r.poolSize -= 1
					// tells everyone including supervisor
					for c := range r.clients {
						c.messageCh <- r.poolSizeMessage()
					}
				}
				// actually deletes client
				delete(r.clients, c)
				if len(r.clients) == 0 {
					r.stop()
					return
				}
			}
		case state := <-r.stateCh:
			r.campaign.State = state
			models.DB.Save(r.campaign)
			for c := range r.clients {
				newMessageState := r.stateMessage(c)
				c.messageCh <- newMessageState
				if newMessageState.Payload == "Unavailable" {
					c.messageCh <- participantDisconnectMessage()
				}
			}
		}
	}
}
