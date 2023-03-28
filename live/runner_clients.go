package live

import (
	"github.com/ducksouplab/mastok/models"
)

type runnerClients struct {
	participants map[*client]bool
	supervisors  map[*client]bool
	all          map[*client]bool
	pool         map[*client]bool
	groupsSize   map[string]int
	groups       map[string]map[*client]bool
}

func newRunnerClients(g *models.Grouping, ps int) *runnerClients {
	groupsSize := make(map[string]int)
	if g == nil {
		// create default group
		groupsSize[defaultGroupLabel] = ps
	} else {
		for _, group := range g.Groups {
			groupsSize[group.Label] = group.Size
		}
	}
	groups := make(map[string]map[*client]bool)
	for label := range groupsSize {
		groups[label] = make(map[*client]bool)
	}
	return &runnerClients{
		participants: make(map[*client]bool),
		supervisors:  make(map[*client]bool),
		all:          make(map[*client]bool),
		pool:         make(map[*client]bool), // have landed, agreed, chosen
		groupsSize:   groupsSize,
		groups:       groups,
	}
}

func (rc *runnerClients) isEmpty() bool {
	return len(rc.all) == 0
}

func (rc *runnerClients) participantsCount() (count int) {
	for _, participants := range rc.groups {
		count += len(participants)
	}
	return
}

func (rc *runnerClients) tentativePool() ([]*client, bool) {
	ok := true
	// all groups have to be full
	for label, size := range rc.groupsSize {
		isGroupFull := len(rc.groups[label]) == size
		ok = ok && isGroupFull
	}
	if ok {
		var flatAgreeingParticipants []*client
		for _, group := range rc.groups {
			for participant := range group {
				flatAgreeingParticipants = append(flatAgreeingParticipants, participant)
			}
		}
		return flatAgreeingParticipants, true
	}
	return nil, false
}

func (rc *runnerClients) has(c *client) bool {
	_, ok := rc.all[c]
	return ok
}

func (rc *runnerClients) add(c *client) {
	rc.all[c] = true
	if c.isSupervisor {
		rc.supervisors[c] = true
	} else {
		rc.participants[c] = true
	}
}

func (rc *runnerClients) choose(c *client, label string) {
	rc.groups[label][c] = true
	rc.pool[c] = true
}

func (rc *runnerClients) delete(c *client) (wasAgreeing bool) {
	delete(rc.all, c)
	delete(rc.pool, c)

	if c.isSupervisor {
		delete(rc.supervisors, c)
	} else {
		delete(rc.participants, c)

		for _, group := range rc.groups {
			_, wasInGroup := group[c]
			delete(group, c)
			wasAgreeing = wasAgreeing || wasInGroup
		}
	}
	return
}
