package live

import (
	"log"
	"strconv"

	"github.com/ducksouplab/mastok/models"
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

func (r *runner) done() chan struct{} {
	return r.doneCh
}

func (r *runner) stop() {
	deleteRunner(r.campaign.Namespace)
	close(r.doneCh)
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
						for client := range r.clients {
							client.signalCh <- "StartSession:" + "url"
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
			log.Printf(">>>>>>>>>>>>>> %v", state)
			for client := range r.clients {
				client.signalCh <- "State:" + state
			}
		}
	}
}
