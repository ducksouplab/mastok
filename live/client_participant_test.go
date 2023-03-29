package live

import (
	"testing"

	"github.com/ducksouplab/mastok/models"
	"github.com/stretchr/testify/assert"
)

func TestClient_Participant_Unit(t *testing.T) {
	t.Run("rejects landing if fingerprint payload is empty", func(t *testing.T) {
		ns := "fxt_par"
		defer tearDown(ns)

		ws := runParticipantStub(ns)
		ws.push(Message{"Land", ""})

		assert.True(t, retryUntil(shortDuration, func() bool {
			_, ok := ws.hasReceivedKind("Reject")
			return ok
		}))
	})

	t.Run("accepts landing if fingerprint payload is present", func(t *testing.T) {
		ns := "fxt_par"
		defer tearDown(ns)

		ws := runParticipantStub(ns)
		ws.push(Message{"Land", "fingerprint"})

		assert.False(t, retryUntil(longDuration, func() bool {
			_, ok := ws.hasReceivedKind("Reject")
			return ok
		}))
	})
}

func TestClient_Participant_Integration(t *testing.T) {

	t.Run("participant receives State first thing", func(t *testing.T) {
		ns := "fxt_par"
		defer tearDown(ns)

		ws := runParticipantStub(ns)

		assert.True(t, retryUntil(shortDuration, func() bool {
			_, ok := ws.hasReceivedKind("State")
			return ok
		}), "participant should receive State")
	})

	t.Run("same fingerprint is rejected from room if campaign requires unique participants", func(t *testing.T) {
		ns := "fxt_par_once"
		defer tearDown(ns)

		// the fixture data is what we expected
		campaign, _ := models.GetCampaignByNamespace(ns)
		assert.Equal(t, true, campaign.JoinOnce)

		ws1 := runParticipantStub(ns)
		ws2 := runParticipantStub(ns)
		ws1.landWith("fingerprint1")
		ws2.landWith("fingerprint1")

		assert.True(t, retryUntil(longDuration, func() bool {
			_, ok := ws2.hasReceivedKind("Reject")
			return ok
		}))
	})

	t.Run("same fingerprint is accepted in room if campaign does not require unique participants", func(t *testing.T) {
		ns := "fxt_par"
		defer tearDown(ns)

		// the fixture data is what we expected
		campaign, _ := models.GetCampaignByNamespace(ns)
		assert.Equal(t, false, campaign.JoinOnce)

		ws1 := runParticipantStub(ns)
		ws2 := runParticipantStub(ns)
		ws1.landWith("fingerprint1")
		ws2.landWith("fingerprint1")

		assert.False(t, retryUntil(longDuration, func() bool {
			_, ok := ws2.hasReceivedKind("Reject")
			return ok
		}))
	})

	t.Run("participant should not receive Consent before landing", func(t *testing.T) {
		ns := "fxt_par"
		defer tearDown(ns)

		ws := runParticipantStub(ns)

		assert.False(t, retryUntil(shortDuration, func() bool {
			_, ok := ws.hasReceivedKind("Consent")
			return ok
		}))
	})

	t.Run("participant receives Consent after landing", func(t *testing.T) {
		ns := "fxt_par"
		defer tearDown(ns)

		ws := runParticipantStub(ns)
		ws.land()

		assert.True(t, retryUntil(shortDuration, func() bool {
			_, ok := ws.hasReceivedKind("Consent")
			return ok
		}))
	})

	t.Run("without grouping, PoolSize is sent after Agree", func(t *testing.T) {
		ns := "fxt_par"
		defer tearDown(ns)

		ws := runParticipantStub(ns)

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
		ns := "fxt_par_paused"
		defer tearDown(ns)

		// the fixture data is what we expected
		campaign, _ := models.GetCampaignByNamespace(ns)
		assert.Equal(t, "Paused", campaign.State)

		ws := runParticipantStub(ns)

		assert.True(t, retryUntil(shortDuration, func() bool {
			ok := ws.hasReceived(Message{"State", "Unavailable"})
			return ok
		}), "participant should receive State:Unavailable")
		assert.True(t, retryUntil(shortDuration, func() bool {
			return ws.isLastWriteKind("Disconnect")
		}), "participant should receive Disconnect")
	})

	t.Run("kicks participants if State is Completed", func(t *testing.T) {
		ns := "fxt_par_completed"
		defer tearDown(ns)

		// the fixture data is what we expected
		campaign, _ := models.GetCampaignByNamespace(ns)
		assert.Equal(t, "Completed", campaign.State)

		ws := runParticipantStub(ns)

		assert.True(t, retryUntil(shortDuration, func() bool {
			ok := ws.hasReceived(Message{"State", "Unavailable"})
			return ok
		}), "participant should receive State:Unavailable")
		assert.True(t, retryUntil(shortDuration, func() bool {
			return ws.isLastWriteKind("Disconnect")
		}), "participant should receive Disconnect")
	})

	t.Run("prevents participant from changing State", func(t *testing.T) {
		ns := "fxt_par"
		defer tearDown(ns)

		// 2 participants
		wsSlice := runParticipantStubs(ns, 2)
		wsSlice[0].push(Message{"State", "Paused"})

		// no one should have received the State update
		for _, ws := range wsSlice {
			assert.False(t, retryUntil(shortDuration, func() bool {
				return ws.hasReceived(Message{"State", "Paused"})
			}), "participant should not receive State:Paused")
		}
	})
}
