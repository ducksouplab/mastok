package live

import (
	"github.com/ducksouplab/mastok/models"
)

type runnerClients struct {
	groupMap                    map[string]int
	participants                map[*client]bool
	supervisors                 map[*client]bool
	all                         map[*client]bool
	pool                        map[*client]bool
	groupedAgreeingParticipants map[string][]*client
}

func newRunnerClients(g *models.Grouping, ps int) *runnerClients {
	gMap := make(map[string]int)
	if g == nil {
		// create default group
		gMap[defaultGroupLabel] = ps
	} else {
		for _, group := range g.Groups {
			gMap[group.Label] = group.Size
		}
	}
	return &runnerClients{
		groupMap:                    gMap,
		participants:                make(map[*client]bool),
		supervisors:                 make(map[*client]bool),
		all:                         make(map[*client]bool),
		pool:                        make(map[*client]bool), // have landed, agreed, chosen
		groupedAgreeingParticipants: make(map[string][]*client),
	}
}

func (rc *runnerClients) isEmpty() bool {
	return len(rc.all) == 0
}

func (rc *runnerClients) participantsCount() (count int) {
	for _, participants := range rc.groupedAgreeingParticipants {
		count += len(participants)
	}
	return
}

func (rc *runnerClients) tentativePool() ([]*client, bool) {
	ok := true
	// all groups have to be full
	for label, size := range rc.groupMap {
		isGroupFull := len(rc.groupedAgreeingParticipants[label]) == size
		ok = ok && isGroupFull
	}
	if ok {
		var flatAgreeingParticipants []*client
		for _, participants := range rc.groupedAgreeingParticipants {
			flatAgreeingParticipants = append(flatAgreeingParticipants, participants...)
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

func (rc *runnerClients) choose(c *client, groupLabel string) {
	rc.groupedAgreeingParticipants[defaultGroupLabel] = append(rc.groupedAgreeingParticipants[defaultGroupLabel], c)
	rc.pool[c] = true
}

func deleteFromSlice(c *client, list []*client) (newList []*client, wasThere bool) {
	for _, item := range list {
		if item == c {
			wasThere = true
		} else {
			newList = append(newList, item)
		}
	}
	return
}

func (rc *runnerClients) delete(c *client) (wasAgreeing bool) {
	delete(rc.all, c)
	delete(rc.pool, c)

	if c.isSupervisor {
		delete(rc.supervisors, c)
	} else {
		delete(rc.participants, c)

		for groupLabel, participants := range rc.groupedAgreeingParticipants {
			newParticipants, wasThere := deleteFromSlice(c, participants)
			// replace
			rc.groupedAgreeingParticipants[groupLabel] = newParticipants
			wasAgreeing = wasAgreeing || wasThere
		}
	}
	return
}
