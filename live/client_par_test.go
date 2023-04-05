package live

import (
	"testing"

	"github.com/ducksouplab/mastok/models"
	th "github.com/ducksouplab/mastok/test_helpers"
	"github.com/stretchr/testify/assert"
)

func TestClient_Participant_Integration(t *testing.T) {

	t.Run("participant receives State first thing", func(t *testing.T) {
		ns := "fxt_par"
		defer tearDown(ns)

		ws := runParticipantStub(ns)

		assert.True(t, retryUntil(shortDuration, func() bool {
			return ws.hasReceivedKind("State")
		}), "participant should receive State")
	})

	t.Run("without grouping, PoolSize is sent after Agree", func(t *testing.T) {
		ns := "fxt_par"
		defer tearDown(ns)

		ws := runParticipantStub(ns)

		assert.False(t, retryUntil(shortDuration, func() bool {
			return ws.hasReceivedKind("PoolSize")
		}))

		ws.land()

		assert.False(t, retryUntil(shortDuration, func() bool {
			return ws.hasReceivedKind("PoolSize")
		}))

		ws.agree()

		assert.True(t, retryUntil(shortDuration, func() bool {
			return ws.hasReceivedKind("PoolSize")
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
		wsSlice[0].send(Message{"State", "Paused"})

		// no one should have received the State update
		for _, ws := range wsSlice {
			assert.False(t, retryUntil(shortDuration, func() bool {
				return ws.hasReceived(Message{"State", "Paused"})
			}), "participant should not receive State:Paused")
		}
	})

	t.Run("turns Campaign to completed after last SessionStart", func(t *testing.T) {
		ns := "fxt_par_almost_completed"
		perSession := 4
		defer tearDown(ns)

		// the fixture data is what we expected
		campaign, _ := models.GetCampaignByNamespace(ns)
		assert.Equal(t, perSession, campaign.PerSession)
		assert.Equal(t, 3, campaign.StartedSessions)
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
		assert.True(t, retryUntil(longDuration, func() bool {
			c, _ := models.GetCampaignByNamespace(ns)
			return c.State == models.Completed && c.StartedSessions == 4
		}), "campaign should be Completed")

		// outer state: new participant can't connect
		wsAdditional := runParticipantStub(ns)
		// no need to land/agree, Completed State will kick participant first thing
		assert.True(t, retryUntil(shortDuration, func() bool {
			return wsAdditional.isLastWriteKind("Disconnect")
		}), "participant should receive Disconnect")
	})
}
