package live

import (
	"github.com/ducksouplab/mastok/models"
)

// not guarded by lock since managed by a runner whose concurrency is dealt with channels
type runnerClients struct {
	// configuration
	sizeByGroup       map[string]int
	maxPendingByGroup map[string]int
	perSession        int
	maxPending        int
	// state
	supervisors  map[*client]bool
	participants map[*client]bool
	all          map[*client]bool            // supervisors and participants: used to broadcast messages
	pool         map[*client]bool            // participants selected for next session
	poolByGroup  map[string]map[*client]bool // same contents as pool, but categorized
	pending      []*client                   // participants (ordered by arrival) for following sessions
}

func newRunnerClients(c *models.Campaign, g *models.Grouping) *runnerClients {
	sizeByGroup := make(map[string]int)
	if g == nil {
		// create default group
		sizeByGroup[defaultGroupLabel] = c.PerSession
	} else {
		for _, group := range g.Groups {
			sizeByGroup[group.Label] = group.Size
		}
	}

	groups := make(map[string]map[*client]bool)
	for label := range sizeByGroup {
		groups[label] = make(map[*client]bool)
	}

	maxPending := 0
	maxPendingByGroup := make(map[string]int)
	margin := (c.MaxSessions - c.StartedSessions) + 1
	for label, size := range sizeByGroup {
		max := size * margin
		maxPendingByGroup[label] = max
		maxPending += max
	}

	return &runnerClients{
		sizeByGroup:       sizeByGroup,
		maxPendingByGroup: maxPendingByGroup,
		maxPending:        maxPending,
		perSession:        c.PerSession,
		supervisors:       make(map[*client]bool),
		participants:      make(map[*client]bool),
		all:               make(map[*client]bool),
		pool:              make(map[*client]bool), // have landed, agreed, chosen
		poolByGroup:       groups,
	}
}

// helpers

func sliceDelete(cSlice []*client, toRemove *client) (newSlice []*client) {
	for _, c := range cSlice {
		if c != toRemove {
			newSlice = append(newSlice, c)
		}
	}
	return
}

func (rc *runnerClients) isGroupFull(label string) bool {
	return len(rc.poolByGroup[label]) == rc.sizeByGroup[label]
}

func (rc *runnerClients) isPendingForGroupFull(label string) bool {
	var pendingForGroupCount int
	for _, c := range rc.pending {
		if c.groupLabel == label {
			pendingForGroupCount++
		}
	}
	return pendingForGroupCount >= rc.maxPendingByGroup[label]
}

// read methods

func (rc *runnerClients) isEmpty() bool {
	return len(rc.all) == 0
}

func (rc *runnerClients) poolSize() (count int) {
	return len(rc.pool)
}

func (rc *runnerClients) pendingSize() (count int) {
	return len(rc.pending)
}

func (rc *runnerClients) isPoolFull() bool {
	return len(rc.pool) == rc.perSession
}

// read-write methods

func (rc *runnerClients) add(c *client) {
	rc.all[c] = true
	if c.isSupervisor {
		rc.supervisors[c] = true
	} else {
		rc.participants[c] = true
	}
}

func (rc *runnerClients) tentativeJoin(c *client) (addedToPool bool, addedToPending bool) {
	if rc.isGroupFull(c.groupLabel) {
		if rc.isPendingForGroupFull(c.groupLabel) {
			return false, false
		} else {
			rc.pending = append(rc.pending, c)
			return false, true
		}
	} else {
		rc.pool[c] = true
		rc.poolByGroup[c.groupLabel][c] = true
		return true, false
	}
}

func (rc *runnerClients) addOneToPoolFromPending() (update bool) {
	// reset pending
	oldPending := make([]*client, len(rc.pending))
	copy(oldPending, rc.pending)
	rc.pending = nil // nil is a valid slice https://github.com/uber-go/guide/blob/master/style.md#nil-is-a-valid-slice

	// fill
	for _, c := range oldPending {
		addedToPool, _ := rc.tentativeJoin(c)
		update = update || addedToPool
	}

	return
}

func (rc *runnerClients) resetPoolFromPending() (update bool) {
	// reset pending
	oldPending := make([]*client, len(rc.pending))
	copy(oldPending, rc.pending)
	rc.pending = nil // nil is a valid slice https://github.com/uber-go/guide/blob/master/style.md#nil-is-a-valid-slice

	// fill
	for _, c := range oldPending {
		addedToPool, _ := rc.tentativeJoin(c)
		update = update || addedToPool
	}

	return
}

func (rc *runnerClients) delete(c *client) (wasInPool bool) {
	delete(rc.all, c)

	if c.isSupervisor {
		delete(rc.supervisors, c)
	} else {
		delete(rc.participants, c)
		delete(rc.pool, c)

		group := rc.poolByGroup[c.groupLabel]
		if _, isInGroup := group[c]; isInGroup {
			delete(group, c)
			wasInPool = true
		} else {
			rc.pending = sliceDelete(rc.pending, c)
		}
	}
	return
}
