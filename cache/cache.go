package cache

import (
	"log"
	"sync"
	"time"

	"github.com/ducksouplab/mastok/env"
	"github.com/ducksouplab/mastok/otree"
)

const TTL = 120

var (
	expCache experimentCache
)

type experiment struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type experimentCache struct {
	sync.Mutex
	updatedAt time.Time
	list      []experiment
	namesMap  map[string]string
}

func init() {
	if env.AsCommandLine {
		// don't load cache in command line mode
		return
	}
	expCache = experimentCache{sync.Mutex{}, time.Now(), nil, make(map[string]string)}
	if env.Mode != "TEST" {
		GetExperiments()
	}
}

func notExpired(t time.Time) bool {
	return time.Since(t).Seconds() < TTL
}

func GetExperimentName(id string) string {
	return expCache.namesMap[id]
}

func GetExperiments() []experiment {
	// use cache
	if expCache.list != nil && notExpired(expCache.updatedAt) {
		return expCache.list
	}
	// or (re)fetch and update cache
	list := []experiment{}
	err := otree.GetOTreeJSON("/api/session_configs", &list)
	if err != nil {
		log.Fatal(err)
	}

	expCache.Lock()
	expCache.list = list
	expCache.namesMap = make(map[string]string)
	for _, e := range list {
		expCache.namesMap[e.Id] = e.Name
	}
	expCache.updatedAt = time.Now()
	expCache.Unlock()

	return list
}
