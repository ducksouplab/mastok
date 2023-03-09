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
	poolSize uint
	clients  map[*client]bool
	// manage broadcasting
	registerCh   chan *client
	unregisterCh chan *client
	// other incoming events
	stateCh    chan string
	joinPoolCh chan *client
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

func (r *runner) sessionStart() (sessionCode string, participantCodes []string, err error) {
	args := otree.SessionArgs{
		ConfigName:      r.campaign.ExperimentConfig,
		NumParticipants: int(r.campaign.PerSession),
		Config: otree.NestedConfig{
			Id: "mk:" + r.campaign.Namespace,
		},
	}
	s := otree.Session{}
	// GET code
	if err = otree.PostOTreeJSON("/api/sessions/", args, &s); err != nil {
		return
	}
	sessionCode = s.Code
	// GET more details (participant codes) and override s
	err = otree.GetOTreeJSON("/api/sessions/"+s.Code, &s)

	for _, p := range s.Participants {
		participantCodes = append(participantCodes, p.Code)
	}
	return
}

func (r *runner) loop() {
	for {
		select {
		case client := <-r.registerCh:
			if r.campaign.State == "Running" || client.isSupervisor {
				r.clients[client] = true
				client.signalCh <- "State:" + r.campaign.State

				if !client.isSupervisor {
					// increases pool
					r.poolSize += 1
					for client := range r.clients {
						client.signalCh <- "PoolSize:" + strconv.FormatUint(uint64(r.poolSize), 10)
					}
					// starts session when pool is full
					if r.poolSize == r.campaign.PerSession {
						sessionCode, participantCodes, err := r.sessionStart()
						if err != nil {
							log.Println("[runner] oTree session creation failed")
						} else {
							participantIndex := 0
							for client := range r.clients {
								if client.isSupervisor {
									client.signalCh <- "SessionStart:" + otree.SupervisorStartURL(sessionCode)
								} else {
									client.signalCh <- "SessionStart:" + otree.ParticipantStartURL(participantCodes[participantIndex])
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
						c.signalCh <- "PoolSize:" + strconv.FormatUint(uint64(r.poolSize), 10)
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
				client.signalCh <- "State:" + state
			}
		}
	}
}
