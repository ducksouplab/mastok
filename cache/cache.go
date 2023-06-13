package cache

import (
	"errors"
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
	configs   []otree.Config
}

func init() {
	if env.AsCommandLine {
		// don't load cache in command line mode
		return
	}
	cache = experimentCache{sync.Mutex{}, time.Now(), nil}
	if env.Mode != "TEST" {
		GetOTreeConfigs()
	}
	log.Printf("[cache] oTree experiments: %+v\n", cache.configs)
}

func notExpired(t time.Time) bool {
	return time.Since(t).Seconds() < TTL
}

func GetOTreeConfigs() []otree.Config {
	// use cache
	if cache.configs != nil && notExpired(cache.updatedAt) {
		return cache.configs
	}
	// or (re)fetch and update cache
	var configs []otree.Config
	err := otree.GetOTreeJSON("/api/session_configs", &configs)
	if err != nil {
		return nil
	}

	cache.Lock()
	cache.configs = configs
	cache.updatedAt = time.Now()
	cache.Unlock()

	return configs
}

func GetOTreeConfig(name string) (config otree.Config, err error) {
	for _, xp := range cache.configs {
		if xp.Name == name {
			return xp, nil
		}
	}
	err = errors.New("not found")
	return
}
