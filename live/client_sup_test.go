package live

import (
	"strconv"
	"testing"

	"github.com/ducksouplab/mastok/models"
	th "github.com/ducksouplab/mastok/test_helpers"
	"github.com/stretchr/testify/assert"
)

func TestClient_Supervisor_Integration(t *testing.T) {
	t.Run("supervisor has runner", func(t *testing.T) {
		ns := "fxt_sup"
		defer tearDown(ns)

		runSupervisorStub(ns)

		_, ok := hasRunner(ns)
		assert.Equal(t, true, ok)
	})

	t.Run("supervisor receives PoolSize", func(t *testing.T) {
		ns := "fxt_sup"
		defer tearDown(ns)

		ws, _ := runSupervisorStub(ns)

		assert.True(t, retryUntil(shortDuration, func() bool {
			return ws.hasReceivedKind("PoolSize")
		}), "supervisor should receive PoolSize")
	})

	t.Run("supervisor receives PendingSize", func(t *testing.T) {
		ns := "fxt_sup_grouping"
		defer tearDown(ns)

		wsSup, campaign := runSupervisorStub(ns)

		// the fixture data is what we expected
		assert.Equal(t, 3, campaign.PerSession)
		assert.Contains(t, campaign.Grouping, "Male:1")
		assert.Contains(t, campaign.Grouping, "Female:1")
		assert.Contains(t, campaign.Grouping, "Other:1")

		// first participants in pool
		ws := runParticipantStub(ns)
		ws.land().agree().connectWithGroup("Female")

		assert.True(t, retryUntil(shortDuration, func() bool {
			return wsSup.hasReceivedKind("PoolSize")
		}))

		// other participants pending
		wsSlice := runParticipantStubs(ns, 2)
		for _, ws := range wsSlice {
			ws.land().agree().connectWithGroup("Female")
		}

		// 3 groups * 1 person per group * (4 sessions + 1)
		maxPendingSize := 15
		assert.True(t, retryUntil(longDuration, func() bool {
			ok := wsSup.hasReceived(Message{"PendingSize", "2/" + strconv.Itoa(maxPendingSize)})
			return ok
		}))

		// one pending leaves
		wsSlice[0].Close()
		wsSup.ClearAllMessages()
		assert.True(t, retryUntil(longDuration, func() bool {
			ok := wsSup.hasReceived(Message{"PendingSize", "1/" + strconv.Itoa(maxPendingSize)})
			return ok
		}))
	})

	t.Run("aborts session when supervisor changes campaign State to paused", func(t *testing.T) {
		ns := "fxt_to_be_paused"
		defer tearDown(ns)

		wsSup, campaign := runSupervisorStub(ns)
		// 3 participants (session won't start)
		wsSlice := runParticipantStubs(ns, 2)
		for _, ws := range wsSlice {
			ws.land().agree()
		}
		// the fixture data is what we expected
		assert.Equal(t, 4, campaign.PerSession)
		assert.Equal(t, "Running", campaign.State)
		// every participants received the new state
		wsSup.send(Message{"State", "Paused"})
		for _, ws := range wsSlice {
			assert.True(t, retryUntil(longDuration, func() bool {
				return ws.hasReceived(Message{"State", "Unavailable"})
			}), "participant should receive State:Unavailable")
		}
		// participants have been disconnected
		assert.True(t, retryUntil(shortDuration, func() bool {
			return wsSup.hasReceived(Message{"PoolSize", "0/4"})
		}), "supervisor should receive PoolSize:0/4")
	})

	t.Run("persists supervisor changed State after runner stopped", func(t *testing.T) {
		ns := "fxt_sup_paused"
		defer tearDown(ns)

		wsSup1, campaign := runSupervisorStub(ns)

		// the fixture data is what we expected
		assert.Equal(t, "Paused", campaign.State)

		// supervisor changes state and quits
		wsSup1.send(Message{"State", "Running"})
		wsSup1.Close()
		runner, _ := hasRunner(ns)
		<-runner.isDone()

		// other supervisor connects
		wsSup2, _ := runSupervisorStub(ns)
		assert.True(t, retryUntil(longDuration, func() bool {
			return wsSup2.hasReceived(Message{"State", "Running"})
		}), "State should be persisted")
	})

	t.Run("manages Busy state", func(t *testing.T) {
		ns := "fxt_sup_busy"
		perSession := 2
		defer tearDown(ns)

		wsSup, campaign := runSupervisorStub(ns)
		// the fixture data is what we expected
		assert.Equal(t, perSession, campaign.PerSession)
		assert.Equal(t, 1, campaign.ConcurrentSessions)
		assert.Equal(t, "Running", campaign.State)

		// fills the room
		th.InterceptOtreePostSession()
		th.InterceptOtreeGetSession()
		defer th.InterceptOff()

		wsSlice := runParticipantStubs(ns, perSession)
		for _, ws := range wsSlice {
			ws.land().agree()
		}

		// assert inner state
		assert.True(t, retryUntil(longerDuration, func() bool {
			return wsSup.hasReceived(Message{"State", "Busy"})
		}), "supervisor should receive Busy")
		assert.True(t, retryUntil(sessionDurationTest*models.SessionDurationUnit, func() bool {
			return wsSup.isLastWrite(Message{"State", "Running"})
		}), "supervisor should receive Running after Busy")
	})
}
