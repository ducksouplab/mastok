package live

import (
	"sync"

	"github.com/ducksouplab/mastok/models"
)

var (
	rs runnerStore
)

type runnerStore struct {
	sync.Mutex
	index map[string]*runner
}

func init() {
	rs = newRunnerStore()
}

func newRunnerStore() runnerStore {
	return runnerStore{sync.Mutex{}, make(map[string]*runner)}
}

func hasRunner(namespace string) (r *runner, ok bool) {
	rs.Lock()
	defer rs.Unlock()

	r, ok = rs.index[namespace]
	return
}

// get existing or initialize
func getRunner(c *models.Campaign) *runner {
	// already running
	if r, ok := hasRunner(c.Namespace); ok {
		return r
	}
	// create runner
	r := newRunner(c)

	rs.Lock()
	rs.index[c.Namespace] = r
	rs.Unlock()

	go r.loop()
	return r
}

func deleteRunner(c *models.Campaign) {
	rs.Lock()
	defer rs.Unlock()

	delete(rs.index, c.Namespace)
}

// API

func UpdateRunner(c *models.Campaign) {
	// only if already running
	if r, ok := hasRunner(c.Namespace); ok {
		go func() {
			r.updateCampaignCh <- c
		}()
	}
}
