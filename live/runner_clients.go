package live

import (
	"github.com/ducksouplab/mastok/models"
)

// no lock since it is only used by a runner
// whose concurrency access is restricted by a global channel select loop
type runnerClients struct {
	// configuration
	sizeByGroup       map[string]int
	maxPendingByGroup map[string]int
	perSession        int
	maxPending        int
	// state
	supervisors    map[*client]bool
	participants   map[*client]bool
	all            map[*client]bool            // supervisors and participants: used to broadcast messages
	joining        map[*client]bool            // participants selected for next session
	joiningByGroup map[string]map[*client]bool // same contents as joining, but categorized
	pendingList    []*client                   // participants (ordered by arrival) for following sessions
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
		joining:           make(map[*client]bool), // have landed, agreed, chosen
		joiningByGroup:    groups,
	}
}

// helpers

func sliceContains(cSlice []*client, target *client) bool {
	for _, c := range cSlice {
		if target == c {
			return true
		}
	}
	return false
}

func sliceDelete(cSlice []*client, toRemove *client) (newSlice []*client) {
	for _, c := range cSlice {
		if c != toRemove {
			newSlice = append(newSlice, c)
		}
	}
	return
}

func (rc *runnerClients) isGroupFull(label string) bool {
	return len(rc.joiningByGroup[label]) == rc.sizeByGroup[label]
}

func (rc *runnerClients) isPendingForGroupFull(label string) bool {
	var pendingForGroupCount int
	for _, c := range rc.pendingList {
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

func (rc *runnerClients) joiningSize() (count int) {
	return len(rc.joining)
}

func (rc *runnerClients) pendingSize() (count int) {
	return len(rc.pendingList)
}

func (rc *runnerClients) isJoiningFull() bool {
	return len(rc.joining) == rc.perSession
}

// read-write methods

// register for websocket communication
func (rc *runnerClients) register(c *client) {
	rc.all[c] = true
	if c.isSupervisor {
		rc.supervisors[c] = true
	} else {
		rc.participants[c] = true
	}
}

func (rc *runnerClients) tentativePending(c *client) (addedToPending bool) {
	if rc.isPendingForGroupFull(c.groupLabel) {
		return false
	} else {
		if sliceContains(rc.pendingList, c) { // don't append twice
			return false
		} else {
			rc.pendingList = append(rc.pendingList, c)
			return true
		}
	}
}

func (rc *runnerClients) tentativeJoin(c *client) (addedToJoining bool) {
	if rc.isGroupFull(c.groupLabel) {
		return false
	}
	rc.joining[c] = true
	rc.joiningByGroup[c.groupLabel][c] = true
	return true
}

func (rc *runnerClients) oneTentativeJoinFromPending() (updated bool) {
	for _, c := range rc.pendingList {
		if rc.tentativeJoin(c) {
			rc.pendingList = sliceDelete(rc.pendingList, c)
			updated = true
			break
		}
	}
	return
}

func (rc *runnerClients) allTentativeJoinFromPending() (updated bool) {
	for _, c := range rc.pendingList {
		if rc.tentativeJoin(c) {
			rc.pendingList = sliceDelete(rc.pendingList, c)
			updated = true
			// update all: no break
		}
	}
	return
}

func (rc *runnerClients) delete(c *client) (wasInJoining bool, wasInPending bool) {
	// it may happen that delete is called twice (starts then quits joining + js client readLoop stop)
	if _, ok := rc.all[c]; !ok {
		return false, false
	}

	delete(rc.all, c)
	if c.isSupervisor {
		delete(rc.supervisors, c)
		return false, false
	} else {
		delete(rc.participants, c)
		delete(rc.joining, c)

		group := rc.joiningByGroup[c.groupLabel]
		if _, isInGroup := group[c]; isInGroup {
			delete(group, c)
			wasInJoining = true
		} else {
			rc.pendingList = sliceDelete(rc.pendingList, c)
		}
		return wasInJoining, !wasInJoining
	}
}
