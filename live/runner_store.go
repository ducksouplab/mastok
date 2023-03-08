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
func getRunner(namespace string) (*runner, error) {
	// already running
	if r, ok := hasRunner(namespace); ok {
		return r, nil
	}
	// load from DB
	campaign, err := models.FindCampaignByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	r := newRunner(campaign)

	rs.Lock()
	rs.index[namespace] = r
	rs.Unlock()

	go r.loop()
	return r, nil
}

func deleteRunner(namespace string) {
	rs.Lock()
	defer rs.Unlock()

	delete(rs.index, namespace)
}
