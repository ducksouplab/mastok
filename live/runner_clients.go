package live

import (
	"github.com/ducksouplab/mastok/models"
)

type runnerClients struct {
	// configuration
	sizeByGroup map[string]int
	// state
	supervisors  map[*client]bool
	participants map[*client]bool
	all          map[*client]bool            // supervisors and participants: used to broadcast messages
	pool         map[*client]bool            // participants selected for next session
	poolByGroup  map[string]map[*client]bool // same contents as pool, but categorized
	pending      []*client                   // participants (ordered by arrival) for following sessions
}

func newRunnerClients(g *models.Grouping, ps int) *runnerClients {
	sizeByGroup := make(map[string]int)
	if g == nil {
		// create default group
		sizeByGroup[defaultGroupLabel] = ps
	} else {
		for _, group := range g.Groups {
			sizeByGroup[group.Label] = group.Size
		}
	}
	groups := make(map[string]map[*client]bool)
	for label := range sizeByGroup {
		groups[label] = make(map[*client]bool)
	}
	return &runnerClients{
		supervisors:  make(map[*client]bool),
		participants: make(map[*client]bool),
		all:          make(map[*client]bool),
		pool:         make(map[*client]bool), // have landed, agreed, chosen
		poolByGroup:  groups,
		sizeByGroup:  sizeByGroup,
	}
}

func (rc *runnerClients) isEmpty() bool {
	return len(rc.all) == 0
}

func (rc *runnerClients) poolSize() (count int) {
	return len(rc.pool)
}

func (rc *runnerClients) pendingSize() (count int) {
	return len(rc.pending)
}

func (rc *runnerClients) add(c *client) {
	rc.all[c] = true
	if c.isSupervisor {
		rc.supervisors[c] = true
	} else {
		rc.participants[c] = true
	}
}

func (rc *runnerClients) isGroupFull(label string) bool {
	return len(rc.poolByGroup[label]) == rc.sizeByGroup[label]
}

func (rc *runnerClients) tentativeAddToPool(c *client) bool {
	if rc.isGroupFull(c.groupLabel) {
		rc.pending = append(rc.pending, c)
		return false
	} else {
		rc.pool[c] = true
		rc.poolByGroup[c.groupLabel][c] = true
		return true
	}
}

func (rc *runnerClients) isPoolReady() bool {
	ok := true
	// all groups have to be full
	for label := range rc.sizeByGroup {
		isFull := rc.isGroupFull(label)
		ok = ok && isFull
	}
	return ok
}

func (rc *runnerClients) fillPoolFromPending() (update bool) {
	if len(rc.pending) == 0 {
		return
	}
	// reset pending
	oldPending := make([]*client, len(rc.pending))
	copy(oldPending, rc.pending)
	rc.pending = []*client{}
	// fill
	for _, c := range oldPending {
		added := rc.tentativeAddToPool(c)
		update = update || added
	}
	return
}

func sliceDelete(cSlice []*client, toRemove *client) (newSlice []*client) {
	for _, c := range cSlice {
		if c != toRemove {
			newSlice = append(newSlice, c)
		}
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
