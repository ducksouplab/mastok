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

type experimentCache struct {
	sync.Mutex
	updatedAt time.Time
	names     []string
}

func init() {
	if env.AsCommandLine {
		// don't load cache in command line mode
		return
	}
	expCache = experimentCache{sync.Mutex{}, time.Now(), nil}
	if env.Mode != "TEST" {
		GetExperiments()
	}
}

func notExpired(t time.Time) bool {
	return time.Since(t).Seconds() < TTL
}

func GetExperiments() []string {
	// use cache
	if expCache.names != nil && notExpired(expCache.updatedAt) {
		return expCache.names
	}
	// or (re)fetch and update cache
	configs := []otree.ExperimentConfig{}
	err := otree.GetOTreeJSON("/api/session_configs", &configs)
	if err != nil {
		log.Fatal(err)
	}

	expCache.Lock()
	var list []string
	for _, config := range configs {
		list = append(list, config.Name)
	}
	expCache.updatedAt = time.Now()
	expCache.Unlock()

	return list
}
