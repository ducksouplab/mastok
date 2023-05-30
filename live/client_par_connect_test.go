package live

import (
	"testing"

	"github.com/ducksouplab/mastok/models"
	th "github.com/ducksouplab/mastok/test_helpers"
	"github.com/stretchr/testify/assert"
)

func TestClient_Participant_Connect_Integration(t *testing.T) {

	t.Run("participant does not receive Grouping before agreeing", func(t *testing.T) {
		ns := "fxt_connect_grouping"
		defer tearDown(ns)

		campaign, _ := models.GetCampaignByNamespace(ns)

		// the fixture data is what we expected
		assert.Contains(t, campaign.Grouping, "Male:2")
		assert.Contains(t, campaign.Grouping, "Female:2")

		ws := runParticipantStub(ns)
		ws.land()

		assert.False(t, retryUntil(shortDuration, func() bool {
			return ws.hasReceivedKind("Grouping")
		}))
	})

	t.Run("participant receives Grouping after agreeing", func(t *testing.T) {
		ns := "fxt_connect_grouping"
		defer tearDown(ns)

		campaign, _ := models.GetCampaignByNamespace(ns)

		// the fixture data is what we expected
		assert.Contains(t, campaign.Grouping, "Male:2")
		assert.Contains(t, campaign.Grouping, "Female:2")

		ws := runParticipantStub(ns)
		ws.land().agree()

		assert.True(t, retryUntil(shortDuration, func() bool {
			return ws.hasReceivedKind("Grouping")
		}))
	})

	t.Run("participant can connect even if campaign is busy", func(t *testing.T) {
		ns := "fxt_connect_pending_busy"
		defer tearDown(ns)

		wsSup, campaign := runSupervisorStub(ns)

		// the fixture data is what we expected
		assert.Equal(t, 4, campaign.PerSession)
		assert.Contains(t, campaign.Grouping, "Male:2")
		assert.Contains(t, campaign.Grouping, "Female:2")

		// oTree interception is needed when SessionStart
		th.InterceptOtreePostSession()
		th.InterceptOtreeGetSession()
		defer th.InterceptOff()

		wsMales := runParticipantStubs(ns, 2)
		wsFemales := runParticipantStubs(ns, 2)
		for _, ws := range wsMales {
			ws.land().agree().connectWithGroup("Male")
		}
		for _, ws := range wsFemales {
			ws.land().agree().connectWithGroup("Female")
		}

		wsMalesToo := runParticipantStubs(ns, 3)
		for _, ws := range wsMalesToo {
			ws.land().agree().connectWithGroup("Male")
		}
		for _, ws := range wsMalesToo {
			assert.False(t, retryUntil(shortDuration, func() bool {
				return ws.hasReceivedKind("Disconnect")
			}))
		}
		assert.True(t, retryUntil(longDuration, func() bool {
			return wsSup.hasReceived(Message{"State", "Busy"})
		}), "supervisor should receive Busy")
	})

	t.Run("participants receive Pending", func(t *testing.T) {
		ns := "fxt_par_connect_full"
		defer tearDown(ns)

		campaign, _ := models.GetCampaignByNamespace(ns)

		// the fixture data is what we expected
		assert.Equal(t, 4, campaign.PerSession)
		assert.Equal(t, campaign.State, "Busy")
		assert.Equal(t, campaign.Grouping, "")
		assert.Equal(t, campaign.PerSession, 4)

		wsParticipants := runParticipantStubs(ns, 8)
		for _, ws := range wsParticipants {
			ws.land().agree()
		}
		wsLast := wsParticipants[len(wsParticipants)-1]

		assert.True(t, retryUntil(shortDuration, func() bool {
			return wsLast.hasReceivedKind("Pending")
		}))

		assert.True(t, retryUntil(shortDuration, func() bool {
			return wsLast.hasReceived(Message{"State", "Busy"})
		}))
	})

	t.Run("participant is moved from Pending to Joining when other quits Joining", func(t *testing.T) {
		ns := "fxt_par_connect_full"
		defer tearDown(ns)

		campaign, _ := models.GetCampaignByNamespace(ns)

		// the fixture data is what we expected
		assert.Equal(t, 4, campaign.PerSession)
		assert.Equal(t, campaign.State, "Busy")
		assert.Equal(t, campaign.Grouping, "")
		assert.Equal(t, campaign.PerSession, 4)

		wsParticipants := runParticipantStubs(ns, 8)
		for _, ws := range wsParticipants {
			ws.land().agree()
		}
		wsFirstInJoining := wsParticipants[0]
		wsFirstInPending := wsParticipants[4]
		wsSecondInPending := wsParticipants[5]

		wsFirstInJoining.Close()

		assert.True(t, retryUntil(shortDuration, func() bool {
			return wsFirstInPending.hasReceivedKind("JoiningSize")
		}))

		assert.False(t, retryUntil(shortDuration, func() bool {
			return wsSecondInPending.hasReceivedKind("JoiningSize")
		}))
	})

	t.Run("participant can't connect if pending is full", func(t *testing.T) {
		ns := "fxt_par_connect_full"
		defer tearDown(ns)

		campaign, _ := models.GetCampaignByNamespace(ns)

		// the fixture data is what we expected
		assert.Equal(t, 4, campaign.PerSession)
		assert.Equal(t, campaign.State, "Busy")
		assert.Equal(t, campaign.Grouping, "")
		assert.Equal(t, campaign.PerSession, 4)
		assert.Equal(t, campaign.MaxSessions, 3)
		assert.Equal(t, campaign.StartedSessions, 1)

		// accepted in pending : (MaxSessions - StartedSessions + 1) * PerSession
		// (3 - 1 + 1) x 4 => 12 in pending (plus 4 in pool)

		wsParticipants := runParticipantStubs(ns, 17)
		for _, ws := range wsParticipants {
			ws.land().agree()
		}
		wsLastAccepted := wsParticipants[len(wsParticipants)-2]
		wsFirstRejected := wsParticipants[len(wsParticipants)-1]

		assert.True(t, retryUntil(shortDuration, func() bool {
			return wsLastAccepted.hasReceivedKind("Pending")
		}))

		assert.True(t, retryUntil(3*longerDuration, func() bool {
			return wsFirstRejected.hasReceived(Message{"Disconnect", "Full"})
		}))
	})

	t.Run("participant receives instructions", func(t *testing.T) {
		ns := "fxt_instructions"
		defer tearDown(ns)

		campaign, _ := models.GetCampaignByNamespace(ns)

		// the fixture data is what we expected
		assert.Contains(t, campaign.Instructions, "Title")

		ws := runParticipantStub(ns)
		ws.land().agree()

		assert.True(t, retryUntil(shortDuration, func() bool {
			return ws.hasReceivedKind("Instructions")
		}))
	})
}
