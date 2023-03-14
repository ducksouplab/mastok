package live

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Integration(t *testing.T) {
	t.Run("has runner", func(t *testing.T) {
		ns := "fxt_live_ns1"
		defer tearDown(ns)

		ws := newWSStub()
		RunSupervisor(ws, ns)

		_, ok := hasRunner(ns)
		assert.Equal(t, true, ok)
	})

	t.Run("receives State", func(t *testing.T) {
		ns := "fxt_live_ns1"
		slug := "fxt_live_ns1_slug"
		defer tearDown(ns)

		ws := newWSStub()
		RunParticipant(ws, slug)

		assert.True(t, retryUntil(shortDuration, func() bool {
			_, ok := ws.hasReceivedKind("State")
			return ok
		}), "participant should receive State")
	})

	t.Run("participant receives PoolSize", func(t *testing.T) {
		ns := "fxt_live_ns1"
		slug := "fxt_live_ns1_slug"
		defer tearDown(ns)

		ws := newWSStub()
		RunParticipant(ws, slug)

		assert.True(t, retryUntil(shortDuration, func() bool {
			_, ok := ws.hasReceivedKind("PoolSize")
			return ok
		}), "participant should receive PoolSize")
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

	t.Run("kicks participants if State is Paused", func(t *testing.T) {
		ns := "fxt_live_ns3_paused"
		slug := "fxt_live_ns3_paused_slug"
		defer tearDown(ns)

		ws := newWSStub()
		p := RunParticipant(ws, slug)

		// the fixture data is what we expected
		assert.Equal(t, "Paused", p.runner.campaign.State)

		assert.True(t, retryUntil(shortDuration, func() bool {
			ok := ws.hasReceived(Message{"State", "Unavailable"})
			return ok
		}), "participant should receive State:Unavailable")
		assert.True(t, retryUntil(shortDuration, func() bool {
			return ws.isLastWriteLike(Message{"Participant", "Disconnect"})
		}), "participant should receive Disconnect")
	})

	t.Run("kicks participants if State is Completed", func(t *testing.T) {
		ns := "fxt_live_ns6_completed"
		slug := "fxt_live_ns6_completed_slug"
		defer tearDown(ns)

		ws := newWSStub()
		p := RunParticipant(ws, slug)

		// the fixture data is what we expected
		assert.Equal(t, "Completed", p.runner.campaign.State)

		assert.True(t, retryUntil(shortDuration, func() bool {
			ok := ws.hasReceived(Message{"State", "Unavailable"})
			return ok
		}), "participant should receive State:Unavailable")
		assert.True(t, retryUntil(shortDuration, func() bool {
			return ws.isLastWriteLike(Message{"Participant", "Disconnect"})
		}), "participant should receive Disconnect")
	})

	t.Run("prevents participant from changing state", func(t *testing.T) {
		ns := "fxt_live_ns1"
		slug := "fxt_live_ns1_slug"
		defer tearDown(ns)

		// 2 participants
		wsSlice := makeWSStubs(2)
		for _, ws := range wsSlice {
			RunParticipant(ws, slug)
		}
		wsSlice[0].push(Message{"State", "Paused"})

		// no one should have received the State update
		for _, ws := range wsSlice {
			assert.False(t, retryUntil(shortDuration, func() bool {
				return ws.hasReceived(Message{"State", "Paused"})
			}), "participant should not receive State:Paused")
		}
	})

	t.Run("aborts session when campaign is paused", func(t *testing.T) {
		ns := "fxt_live_ns2_to_be_paused"
		slug := "fxt_live_ns2_to_be_paused_slug"
		defer tearDown(ns)

		// 1 supervisor
		supWs := newWSStub()
		RunSupervisor(supWs, ns)
		// 3 participants (session won't start)
		wsSlice := makeWSStubs(3)
		for _, ws := range wsSlice {
			RunParticipant(ws, slug)
		}
		// the fixture data is what we expected
		runner, _ := hasRunner(ns)
		assert.Equal(t, 4, runner.campaign.PerSession)
		assert.Equal(t, "Running", runner.campaign.State)
		// every participants received the new state
		supWs.push(Message{"State", "Paused"})
		for _, ws := range wsSlice {
			assert.True(t, retryUntil(longerDuration, func() bool {
				return ws.hasReceived(Message{"State", "Unavailable"})
			}), "participant should receive State:Unavailable")
		}
		// participants have been disconnected
		assert.True(t, retryUntil(shortDuration, func() bool {

			return supWs.isLastWriteLike(Message{"PoolSize", "0/4"})
		}), "supervisor should receive PoolSize:0/4")
	})

	t.Run("persists State after runner stopped", func(t *testing.T) {
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
		assert.True(t, retryUntil(longerDuration, func() bool {
			return supWs2.hasReceived(Message{"State", "Running"})
		}), "State should be persisted")
	})
}
