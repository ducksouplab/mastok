package live

import (
	"encoding/json"
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
	deleteRunner(r.campaign.Namespace)
	close(r.doneCh)
}

func (r *runner) stateSignal() string {
	return "State:" + r.campaign.State
}

func (r *runner) poolSizeSignal() string {
	return "PoolSize:" + strconv.Itoa(r.poolSize) + "/" + strconv.Itoa(r.campaign.PerSession)
}

func (r *runner) sessionStartParticipantSignal(code string) string {
	return "SessionStart:" + otree.ParticipantStartURL(code)
}

func (r *runner) sessionStartSupervisorSignal(session models.Session) string {
	sessionMsh, _ := json.Marshal(session)
	return "SessionStart:" + string(sessionMsh)
}

func (r *runner) loop() {
	for {
		select {
		case client := <-r.registerCh:
			if r.campaign.State == "Running" || client.isSupervisor {
				r.clients[client] = true
				client.signalCh <- r.stateSignal()

				if !client.isSupervisor {
					// increases pool
					r.poolSize += 1
					for client := range r.clients {
						client.signalCh <- r.poolSizeSignal()
					}
					// starts session when pool is full
					if r.poolSize == r.campaign.PerSession {
						session, participantCodes, err := models.NewSession(r.campaign)
						if err != nil {
							log.Println("[runner] oTree session creation failed")
						} else {
							participantIndex := 0
							for client := range r.clients {
								if client.isSupervisor {
									client.signalCh <- r.sessionStartSupervisorSignal(session)
								} else {
									client.signalCh <- r.sessionStartParticipantSignal(participantCodes[participantIndex])
									participantIndex++
								}
							}
						}
					}
				}
			} else {
				client.signalCh <- "Participant:Disconnect"
				if len(r.clients) == 0 {
					r.stop()
					return
				}
			}
		case client := <-r.unregisterCh:
			if _, ok := r.clients[client]; ok {
				// leaves pool if participant
				if !client.isSupervisor {
					r.poolSize -= 1
					// tells everyone including supervisor
					for c := range r.clients {
						c.signalCh <- r.poolSizeSignal()
					}
				}
				// actually deletes client
				delete(r.clients, client)
				if len(r.clients) == 0 {
					r.stop()
					return
				}
			}
		case state := <-r.stateCh:
			r.campaign.State = state
			models.DB.Save(r.campaign)
			for client := range r.clients {
				client.signalCh <- r.stateSignal()
			}
		}
	}
}
