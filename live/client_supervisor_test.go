package live

import (
	"testing"

	"github.com/ducksouplab/mastok/models"
	th "github.com/ducksouplab/mastok/test_helpers"
	"github.com/stretchr/testify/assert"
)

func TestClientSupervisor_Integration(t *testing.T) {
	t.Run("supervisor has runner", func(t *testing.T) {
		ns := "fxt_live_ns1"
		defer tearDown(ns)

		ws := newWSStub()
		RunSupervisor(ws, ns)

		_, ok := hasRunner(ns)
		assert.Equal(t, true, ok)
	})

	t.Run("supervisor receives PoolSize", func(t *testing.T) {
		ns := "fxt_live_ns1"
		defer tearDown(ns)

		ws := newWSStub()
		RunSupervisor(ws, ns)

		assert.True(t, retryUntil(shortDuration, func() bool {
			_, ok := ws.hasReceivedKind("PoolSize")
			return ok
		}), "supervisor should receive PoolSize")
	})

	t.Run("aborts session when supervisor changes campaign State to paused", func(t *testing.T) {
		ns := "fxt_live_ns2_to_be_paused"
		slug := ns + "_slug"
		defer tearDown(ns)

		// 1 supervisor
		supWs := newWSStub()
		RunSupervisor(supWs, ns)
		// 3 participants (session won't start)
		wsSlice := makeWSStubs(3)
		for _, ws := range wsSlice {
			RunParticipant(ws, slug)
			ws.land().join()
		}
		// the fixture data is what we expected
		runner, _ := hasRunner(ns)
		assert.Equal(t, 4, runner.campaign.PerSession)
		assert.Equal(t, "Running", runner.campaign.State)
		// every participants received the new state
		supWs.push(Message{"State", "Paused"})
		for _, ws := range wsSlice {
			assert.True(t, retryUntil(longDuration, func() bool {
				return ws.hasReceived(Message{"State", "Unavailable"})
			}), "participant should receive State:Unavailable")
		}
		// participants have been disconnected
		assert.True(t, retryUntil(shortDuration, func() bool {
			return supWs.hasReceived(Message{"PoolSize", "0/4"})
		}), "supervisor should receive PoolSize:0/4")
	})

	t.Run("persists supervisor changed State after runner stopped", func(t *testing.T) {
		ns := "fxt_live_ns4_paused"
		defer tearDown(ns)

		supWs1 := newWSStub()
		s := RunSupervisor(supWs1, ns)
		// the fixture data is what we expected
		assert.Equal(t, "Paused", s.runner.campaign.State)

		// supervisor changes state and quits
		supWs1.push(Message{"State", "Running"})
		supWs1.Close()
		<-s.runner.isDone()

		// other supervisor connects
		supWs2 := newWSStub()
		RunSupervisor(supWs2, ns)
		assert.True(t, retryUntil(longDuration, func() bool {
			return supWs2.hasReceived(Message{"State", "Running"})
		}), "State should be persisted")
	})

	t.Run("manages Busy state", func(t *testing.T) {
		ns := "fxt_live_ns8_busy"
		slug := ns + "_slug"
		perSession := 2
		defer tearDown(ns)

		// supervisor
		wsSup := newWSStub()
		s := RunSupervisor(wsSup, ns)
		// the fixture data is what we expected
		campaign := s.runner.campaign
		assert.Equal(t, perSession, campaign.PerSession)
		assert.Equal(t, 1, campaign.ConcurrentSessions)
		assert.Equal(t, "Running", campaign.State)

		// fills the pool
		th.InterceptOtreePostSession()
		th.InterceptOtreeGetSession()
		defer th.InterceptOff()
		wsSlice := makeWSStubs(perSession)
		for _, ws := range wsSlice {
			RunParticipant(ws, slug)
			ws.land().join()
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