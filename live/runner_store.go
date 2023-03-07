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

func getRunner(namespace string) (*runner, error) {
	// already running
	if r, ok := rs.index[namespace]; ok {
		return r, nil
	}
	// load from DB
	campaign, err := models.FindCampaignByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	r := newRunner(campaign)
	rs.index[namespace] = r
	go r.loop()
	return r, nil
}
