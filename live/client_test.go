package live

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Integration(t *testing.T) {
	t.Run("has runner", func(t *testing.T) {
		ns := "fixture_ns1"
		defer tearDown(ns)

		ws := newWSStub()
		RunSupervisor(ws, ns)

		_, ok := hasRunner(ns)
		assert.Equal(t, true, ok)
	})

	t.Run("receives state", func(t *testing.T) {
		ns := "fixture_ns1"
		defer tearDown(ns)

		ws := newWSStub()
		RunParticipant(ws, ns)

		assert.True(t, retryUntil(shortDuration, func() bool {
			_, ok := ws.hasReceivedPrefix("State:")
			return ok
		}), "participant should receive State")
	})

	t.Run("kicks participants if State is not Running", func(t *testing.T) {
		ns := "fixture_ns3_waiting"
		defer tearDown(ns)

		ws := newWSStub()
		p := RunParticipant(ws, ns)

		// the fixture data is what we expected
		assert.Equal(t, "Waiting", p.runner.campaign.State)

		assert.True(t, retryUntil(shortDuration, func() bool {
			return ws.lastWrite() == "Participant:Disconnect"
		}), "participant should receive Disconnect")
	})

	t.Run("cleans up runner when closed", func(t *testing.T) {
		ns := "fixture_ns1"
		// no teardown since we are actually testing the effects of quitting (ws.Close())

		// two clients
		ws1 := newWSStub()
		ws2 := newWSStub()
		RunSupervisor(ws1, ns)
		RunParticipant(ws2, ns)
		// both clients are here
		sharedRunner, ok := hasRunner(ns)
		assert.Equal(t, true, ok)

		// one quits
		ws1.Close()
		_, ok = hasRunner(ns)
		assert.Equal(t, true, ok)

		// the other quits
		ws2.Close()
		<-sharedRunner.isDone()
		_, ok = hasRunner(ns)
		assert.Equal(t, false, ok)
	})

	t.Run("prevents participant from changing state", func(t *testing.T) {
		ns := "fixture_ns1"
		defer tearDown(ns)

		// 2 participants
		wsSlice := makeWSStubs(2)
		for _, ws := range wsSlice {
			RunParticipant(ws, ns)
		}
		wsSlice[0].push("State:Paused")

		// no one should have received the State update
		for _, ws := range wsSlice {
			assert.False(t, retryUntil(shortDuration, func() bool {
				return ws.hasReceived("State:Paused")
			}), "participant should not receive State:Paused")
		}
	})

	t.Run("aborts session when campaign is paused", func(t *testing.T) {
		ns := "fixture_ns2_to_be_paused"
		defer tearDown(ns)

		// 1 supervisor
		supWs := newWSStub()
		RunSupervisor(supWs, ns)
		// 3 participants (session won't start)
		wsSlice := makeWSStubs(3)
		for _, ws := range wsSlice {
			RunParticipant(ws, ns)
		}
		// the fixture data is what we expected
		runner, _ := hasRunner(ns)
		assert.Equal(t, uint(4), runner.campaign.PerSession)
		assert.Equal(t, "Running", runner.campaign.State)
		// every participants received the new state
		supWs.push("State:Paused")
		for _, ws := range wsSlice {
			assert.True(t, retryUntil(longerDuration, func() bool {
				return ws.hasReceived("State:Paused")
			}), "participant should receive State:Paused")
		}
		// participants have been disconnected
		assert.True(t, retryUntil(shortDuration, func() bool {
			return supWs.lastWrite() == "PoolSize:0"
		}), "supervisor should receive PoolSize:0")
	})

	t.Run("persists State after runner stopped", func(t *testing.T) {
		ns := "fixture_ns4_waiting"
		defer tearDown(ns)

		supWs1 := newWSStub()
		s := RunSupervisor(supWs1, ns)
		// the fixture data is what we expected
		assert.Equal(t, "Waiting", s.runner.campaign.State)

		// supervisor changes state and quits
		supWs1.push("State:Running")
		supWs1.Close()
		<-s.runner.isDone()

		// other supervisor connects
		supWs2 := newWSStub()
		RunSupervisor(supWs2, ns)
		assert.True(t, retryUntil(longerDuration, func() bool {
			return supWs2.hasReceived("State:Running")
		}), "State should be persisted")
	})
}
