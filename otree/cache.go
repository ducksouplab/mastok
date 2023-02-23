package otree

import (
	"log"
	"sync"
	"time"
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
}

func notExpired(t time.Time) bool {
	return time.Since(t).Seconds() < TTL
}

func GetExperimentCache() []experiment {
	// use cache
	if eCache.list != nil && notExpired(eCache.updatedAt) {
		return eCache.list
	}
	// or (re)fetch and update cache
	list := []experiment{}
	err := GetOTreeJSON("/api/session_configs", &list)
	if err != nil {
		log.Fatal(err)
	}

	eCache.Lock()
	eCache.list = list
	eCache.updatedAt = time.Now()
	eCache.Unlock()

	return list
}
