package live

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_Unit(t *testing.T) {
	t.Run("rejects landing if fingerprint payload is empty", func(t *testing.T) {
		ns := "fxt_live_ns1"
		slug := ns + "_slug"
		defer tearDown(ns)

		ws := newWSStub()
		p := RunParticipant(ws, slug)
		ws.push(Message{"Land", ""})

		time.Sleep(shortDuration)
		assert.False(t, p.hasLanded)
	})

	t.Run("accepts landing if fingerprint payload is present", func(t *testing.T) {
		ns := "fxt_live_ns1"
		slug := ns + "_slug"
		defer tearDown(ns)

		ws := newWSStub()
		p := RunParticipant(ws, slug)
		ws.push(Message{"Land", "fingerprint"})

		time.Sleep(shortDuration)
		assert.True(t, p.hasLanded)
	})
}

func TestClient_Integration(t *testing.T) {
	t.Run("participant receives State first thing", func(t *testing.T) {
		ns := "fxt_live_ns1"
		slug := ns + "_slug"
		defer tearDown(ns)

		ws := newWSStub()
		RunParticipant(ws, slug)

		assert.True(t, retryUntil(shortDuration, func() bool {
			_, ok := ws.hasReceivedKind("State")
			return ok
		}), "participant should receive State")
	})

	t.Run("participant should not receive PoolSize before landing", func(t *testing.T) {
		ns := "fxt_live_ns1"
		slug := ns + "_slug"
		defer tearDown(ns)

		ws := newWSStub()
		RunParticipant(ws, slug)

		assert.False(t, retryUntil(shortDuration, func() bool {
			_, ok := ws.hasReceivedKind("PoolSize")
			return ok
		}))
	})

	t.Run("participant should not receive PoolSize before joining", func(t *testing.T) {
		ns := "fxt_live_ns1"
		slug := ns + "_slug"
		defer tearDown(ns)

		ws := newWSStub()
		RunParticipant(ws, slug)
		ws.land()

		assert.False(t, retryUntil(shortDuration, func() bool {
			_, ok := ws.hasReceivedKind("PoolSize")
			return ok
		}))
	})

	t.Run("participant receives PoolSize after joining", func(t *testing.T) {
		ns := "fxt_live_ns1"
		slug := ns + "_slug"
		defer tearDown(ns)

		ws := newWSStub()
		RunParticipant(ws, slug)
		ws.land().join()

		assert.True(t, retryUntil(shortDuration, func() bool {
			_, ok := ws.hasReceivedKind("PoolSize")
			return ok
		}), "participant should receive PoolSize")
	})

	t.Run("kicks participants if State is Paused", func(t *testing.T) {
		ns := "fxt_live_ns3_paused"
		slug := ns + "_slug"
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
			return ws.isLastWriteKind("Disconnect")
		}), "participant should receive Disconnect")
	})

	t.Run("kicks participants if State is Completed", func(t *testing.T) {
		ns := "fxt_live_ns6_completed"
		slug := ns + "_slug"
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
			return ws.isLastWriteKind("Disconnect")
		}), "participant should receive Disconnect")
	})

	t.Run("prevents participant from changing State", func(t *testing.T) {
		ns := "fxt_live_ns1"
		slug := ns + "_slug"
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
}
