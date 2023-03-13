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
	eCache experimentCache
)

type experiment struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type experimentCache struct {
	sync.Mutex
	updatedAt time.Time
	list      []experiment
}

func init() {
	eCache = experimentCache{sync.Mutex{}, time.Now(), nil}
	if env.Mode != "TEST" {
		GetSessions()
	}
}

func notExpired(t time.Time) bool {
	return time.Since(t).Seconds() < TTL
}

func GetSessions() []experiment {
	// use cache
	if eCache.list != nil && notExpired(eCache.updatedAt) {
		return eCache.list
	}
	// or (re)fetch and update cache
	list := []experiment{}
	err := otree.GetOTreeJSON("/api/session_configs", &list)
	if err != nil {
		log.Fatal(err)
	}

	eCache.Lock()
	eCache.list = list
	eCache.updatedAt = time.Now()
	eCache.Unlock()

	return list
}