package live

import (
	"sync"

	"github.com/ducksouplab/mastok/models"
)

type runner struct {
	sync.Mutex
	campaign *models.Campaign
	clients  map[*client]bool
	// manage broadcasting
	registerCh   chan *client
	unregisterCh chan *client
	// actual events
	stateCh      chan string
	joinCh       chan string
	newSessionCh chan string
}

func newRunner(c *models.Campaign) *runner {
	return &runner{
		sync.Mutex{},
		c,
		make(map[*client]bool),
		make(chan *client),
		make(chan *client),
		make(chan string),
		make(chan string),
		make(chan string),
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
				close(client.signal)
			}
		case state := <-r.stateCh:
			for client := range r.clients {
				client.signal <- state
			}
		}
	}
}
