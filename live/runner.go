package live

import (
	"github.com/ducksouplab/mastok/models"
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
	// actual events
	updateStateCh chan string
	joinPoolCh    chan *client
	leavePoolCh   chan *client
	newSessionCh  chan string
}

func newRunner(c *models.Campaign) *runner {
	return &runner{
		campaign:      c,
		poolSize:      0,
		clients:       make(map[*client]bool),
		registerCh:    make(chan *client),
		unregisterCh:  make(chan *client),
		updateStateCh: make(chan string),
		joinPoolCh:    make(chan *client),
		leavePoolCh:   make(chan *client),
		newSessionCh:  make(chan string),
	}
}

func (r *runner) loop() {
	for {
		select {
		case client := <-r.registerCh:
			r.clients[client] = true
		case client := <-r.unregisterCh:
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				if len(r.clients) == 0 {
					deleteRunner(r.campaign.Namespace)
					return
				}
			}
		case state := <-r.updateStateCh:
			for client := range r.clients {
				client.stateCh <- state
			}
		case <-r.joinPoolCh:
			r.poolSize += 1
			for client := range r.clients {
				client.poolSizeCh <- r.poolSize
			}
		case <-r.leavePoolCh:
			r.poolSize -= 1
			for client := range r.clients {
				client.poolSizeCh <- r.poolSize
			}
		}
	}
}
