package live

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Participant_Unit(t *testing.T) {
	t.Run("rejects landing if fingerprint payload is empty", func(t *testing.T) {
		ns := "fxt_live_ns1"
		slug := ns + "_slug"
		defer tearDown(ns)

		ws := newWSStub()
		RunParticipant(ws, slug)
		ws.push(Message{"Land", ""})

		assert.True(t, retryUntil(shortDuration, func() bool {
			_, ok := ws.hasReceivedKind("Reject")
			return ok
		}))
	})

	t.Run("accepts landing if fingerprint payload is present", func(t *testing.T) {
		ns := "fxt_live_ns1"
		slug := ns + "_slug"
		defer tearDown(ns)

		ws := newWSStub()
		RunParticipant(ws, slug)
		ws.push(Message{"Land", "fingerprint"})

		assert.False(t, retryUntil(longDuration, func() bool {
			_, ok := ws.hasReceivedKind("Reject")
			return ok
		}))
	})
}

func TestClient_Participant_Integration(t *testing.T) {

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

	t.Run("same fingerprint is rejected from room if campaign requires unique participants", func(t *testing.T) {
		ns := "fxt_live_par_once"
		slug := ns + "_slug"
		defer tearDown(ns)

		ws1 := newWSStub()
		p1 := RunParticipant(ws1, slug)
		ws2 := newWSStub()
		RunParticipant(ws2, slug)

		// the fixture data is what we expected
		assert.Equal(t, true, p1.runner.campaign.JoinOnce)

		ws1.landWith("fingerprint1")
		ws2.landWith("fingerprint1")

		assert.True(t, retryUntil(longDuration, func() bool {
			_, ok := ws2.hasReceivedKind("Reject")
			return ok
		}))
	})

	t.Run("same fingerprint is accepted in room if campaign does not require unique participants", func(t *testing.T) {
		ns := "fxt_live_ns1"
		slug := ns + "_slug"
		defer tearDown(ns)

		ws1 := newWSStub()
		p1 := RunParticipant(ws1, slug)
		ws2 := newWSStub()
		RunParticipant(ws2, slug)

		// the fixture data is what we expected
		assert.Equal(t, false, p1.runner.campaign.JoinOnce)

		ws1.landWith("fingerprint1")
		ws2.landWith("fingerprint1")

		assert.False(t, retryUntil(longDuration, func() bool {
			_, ok := ws2.hasReceivedKind("Reject")
			return ok
		}))
	})

	t.Run("participant should not receive Consent before landing", func(t *testing.T) {
		ns := "fxt_live_ns1"
		slug := ns + "_slug"
		defer tearDown(ns)

		ws := newWSStub()
		RunParticipant(ws, slug)

		assert.False(t, retryUntil(shortDuration, func() bool {
			_, ok := ws.hasReceivedKind("Consent")
			return ok
		}))
	})

	t.Run("participant receives Consent after landing", func(t *testing.T) {
		ns := "fxt_live_ns1"
		slug := ns + "_slug"
		defer tearDown(ns)

		ws := newWSStub()
		RunParticipant(ws, slug)
		ws.land()

		assert.True(t, retryUntil(shortDuration, func() bool {
			_, ok := ws.hasReceivedKind("Consent")
			return ok
		}))
	})

	t.Run("without grouping, PoolSize is sent after Agree", func(t *testing.T) {
		ns := "fxt_live_ns1"
		slug := ns + "_slug"
		defer tearDown(ns)

		ws := newWSStub()
		RunParticipant(ws, slug)

		assert.False(t, retryUntil(shortDuration, func() bool {
			_, ok := ws.hasReceivedKind("PoolSize")
			return ok
		}))

		ws.land()

		assert.False(t, retryUntil(shortDuration, func() bool {
			_, ok := ws.hasReceivedKind("PoolSize")
			return ok
		}))

		ws.agree()

		assert.True(t, retryUntil(shortDuration, func() bool {
			_, ok := ws.hasReceivedKind("PoolSize")
			return ok
		}), "participant should receive PoolSize")
	})

	t.Run("kicks participants if State is Paused", func(t *testing.T) {
		ns := "fxt_live_par_paused"
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
		ns := "fxt_live_par_completed"
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
