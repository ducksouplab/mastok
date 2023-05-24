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
	cache experimentCache
)

type experimentCache struct {
	sync.Mutex
	updatedAt time.Time
	configs   []otree.ExperimentConfig
}

func init() {
	if env.AsCommandLine {
		// don't load cache in command line mode
		return
	}
	cache = experimentCache{sync.Mutex{}, time.Now(), nil}
	if env.Mode != "TEST" {
		GetExperiments()
	}
}

func notExpired(t time.Time) bool {
	return time.Since(t).Seconds() < TTL
}

func GetExperiments() []otree.ExperimentConfig {
	// use cache
	if cache.configs != nil && notExpired(cache.updatedAt) {
		return cache.configs
	}
	// or (re)fetch and update cache
	var configs []otree.ExperimentConfig
	err := otree.GetOTreeJSON("/api/session_configs", &configs)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	cache.Lock()
	cache.configs = configs
	cache.updatedAt = time.Now()
	cache.Unlock()

	return configs
}
