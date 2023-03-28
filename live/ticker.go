package live

import (
	"sync"
	"time"

	"github.com/ducksouplab/mastok/models"
)

type ticker struct {
	sync.Mutex
	*time.Ticker
	doneCh    chan struct{}
	isStopped bool
}

func (t *ticker) stop() {
	t.Lock()
	defer t.Unlock()

	if !t.isStopped {
		t.isStopped = true
		t.Ticker.Stop()
		close(t.doneCh)
	}
}

func (t *ticker) loop(r *runner) {
	for {
		select {
		case <-r.doneCh:
			return
		case <-r.updateStateTicker.doneCh:
			return
		case <-r.updateStateTicker.Ticker.C:
			if r.campaign.GetPublicState(true) != models.Busy {
				for c := range r.clients.all {
					if c.isSupervisor {
						c.outgoingCh <- stateMessage(r.campaign, c)
					}
				}
				r.updateStateTicker.stop()
			}
		}
	}
}

func newTicker(d time.Duration) *ticker {
	return &ticker{
		sync.Mutex{},
		time.NewTicker(d),
		make(chan struct{}),
		false,
	}
}
