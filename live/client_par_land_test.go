package live

import (
	"fmt"
	"testing"
	"time"

	"github.com/ducksouplab/mastok/models"
	th "github.com/ducksouplab/mastok/test_helpers"
	"github.com/stretchr/testify/assert"
)

func TestClient_Participant_Unit(t *testing.T) {
	t.Run("rejects landing if fingerprint payload is empty", func(t *testing.T) {
		ns := "fxt_par"
		defer tearDown(ns)

		ws := runParticipantStub(ns)
		ws.send(Message{"Land", ""})

		assert.True(t, retryUntil(shortDuration, func() bool {
			ok := ws.hasReceived(Message{"Disconnect", "LandingFailed"})
			return ok
		}))
	})

	t.Run("accepts landing if fingerprint payload is present", func(t *testing.T) {
		ns := "fxt_par"
		defer tearDown(ns)

		ws := runParticipantStub(ns)
		ws.send(Message{"Land", "fingerprint"})

		assert.False(t, retryUntil(longDuration, func() bool {
			ok := ws.hasReceived(Message{"Disconnect", "LandingFailed"})
			return ok
		}))
	})
}

func TestClient_Participant_Landing_Integration(t *testing.T) {

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
			ok := ws2.hasReceived(Message{"Disconnect", "LandingFailed"})
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
			ok := ws2.hasReceived(Message{"Disconnect", "LandingFailed"})
			return ok
		}))
	})

	t.Run("participant does not receive Consent before landing", func(t *testing.T) {
		ns := "fxt_par"
		defer tearDown(ns)

		ws := runParticipantStub(ns)

		assert.False(t, retryUntil(shortDuration, func() bool {
			return ws.hasReceivedKind("Consent")
		}))
	})

	t.Run("participant receives Consent after landing", func(t *testing.T) {
		ns := "fxt_par"
		defer tearDown(ns)

		ws := runParticipantStub(ns)
		ws.land()

		assert.True(t, retryUntil(shortDuration, func() bool {
			return ws.hasReceivedKind("Consent")
		}))
	})

	t.Run("sends redirect if participant reconnects", func(t *testing.T) {
		ns := "fxt_par_redirect"
		perSession := 2
		defer tearDown(ns)

		// the fixture data is what we expected
		campaign, _ := models.GetCampaignByNamespace(ns)
		assert.Equal(t, perSession, campaign.PerSession)
		assert.Equal(t, false, campaign.JoinOnce)
		assert.Equal(t, 0, campaign.StartedSessions)
		assert.Equal(t, "Running", campaign.State)

		// fills the room
		th.InterceptOtreePostSession()
		th.InterceptOtreeGetSession()
		defer th.InterceptOff()

		// complete room with 2 participants
		wsSlice := runParticipantStubs(ns, perSession)
		for index, ws := range wsSlice {
			ws.landWith(fmt.Sprintf("fingerprint%v", index)).agree()
		}

		// assert session has started
		for _, ws := range wsSlice {
			assert.True(t, retryUntil(longDuration, func() bool {
				return ws.hasReceivedKind("Starting")
			}), "participant should receive Starting with oTree starting link")
		}

		// first participant reconnects (same fingerprint)
		wsSlice[0].Close()
		ws := runParticipantStub(ns)
		ws.landWith("fingerprint0").agree()

		assert.True(t, retryUntil(longDuration, func() bool {
			ok := ws.hasReceivedWithPayloadPrefix(Message{"Disconnect", "Redirect:"})
			return ok
		}), "participant should receive Redirect")
	})

	t.Run("sends redirect if participant reconnects even for JoinOnce campaign", func(t *testing.T) {
		ns := "fxt_par_redirect2"
		perSession := 2
		defer tearDown(ns)

		// the fixture data is what we expected
		campaign, _ := models.GetCampaignByNamespace(ns)
		assert.Equal(t, perSession, campaign.PerSession)
		assert.Equal(t, true, campaign.JoinOnce)
		assert.Equal(t, 0, campaign.StartedSessions)
		assert.Equal(t, "Running", campaign.State)

		// fills the room
		th.InterceptOtreePostSession()
		th.InterceptOtreeGetSession()
		defer th.InterceptOff()

		// complete room with 2 participants
		wsSlice := runParticipantStubs(ns, perSession)
		for index, ws := range wsSlice {
			ws.landWith(fmt.Sprintf("fingerprint%v", index)).agree()
		}

		// assert session has started
		for _, ws := range wsSlice {
			assert.True(t, retryUntil(longDuration, func() bool {
				return ws.hasReceivedKind("Starting")
			}), "participant should receive Starting with oTree starting link")
		}

		// first participant reconnects (same fingerprint)
		wsSlice[0].Close()
		ws := runParticipantStub(ns)
		ws.landWith("fingerprint0").agree()

		assert.True(t, retryUntil(longDuration, func() bool {
			ok := ws.hasReceivedWithPayloadPrefix(Message{"Disconnect", "Redirect:"})
			return ok
		}), "participant should receive Redirect")
	})

	t.Run("sends reject if participant reconnect to JoinOnce campaign after session ended", func(t *testing.T) {
		ns := "fxt_par_reject"
		perSession := 2
		defer tearDown(ns)

		// the fixture data is what we expected
		campaign, _ := models.GetCampaignByNamespace(ns)
		assert.Equal(t, perSession, campaign.PerSession)
		assert.Equal(t, true, campaign.JoinOnce)
		assert.Equal(t, 0, campaign.StartedSessions)
		assert.Equal(t, "Running", campaign.State)

		// fills the room
		th.InterceptOtreePostSession()
		th.InterceptOtreeGetSession()
		defer th.InterceptOff()

		// complete room with 2 participants
		wsSlice := runParticipantStubs(ns, perSession)
		for index, ws := range wsSlice {
			ws.landWith(fmt.Sprintf("fingerprint%v", index)).agree()
		}

		// assert session has started
		for _, ws := range wsSlice {
			assert.True(t, retryUntil(longDuration, func() bool {
				return ws.hasReceivedKind("Starting")
			}), "participant should receive Starting with oTree starting link")
		}

		// first participant reconnects (same fingerprint)
		time.Sleep((sessionDurationTest + 1) * models.SessionDurationUnit)
		ws := runParticipantStub(ns)
		ws.landWith("fingerprint0").agree()

		assert.True(t, retryUntil(longDuration, func() bool {
			ok := ws.hasReceived(Message{"Disconnect", "LandingFailed"})
			return ok
		}))
	})

	t.Run("does not send reject if participant reconnect to multi-join campaign after session ended", func(t *testing.T) {
		ns := "fxt_par_noreject"
		perSession := 2
		defer tearDown(ns)

		// the fixture data is what we expected
		campaign, _ := models.GetCampaignByNamespace(ns)
		assert.Equal(t, perSession, campaign.PerSession)
		assert.Equal(t, false, campaign.JoinOnce)
		assert.Equal(t, 0, campaign.StartedSessions)
		assert.Equal(t, "Running", campaign.State)

		// fills the room
		th.InterceptOtreePostSession()
		th.InterceptOtreeGetSession()
		defer th.InterceptOff()

		// complete room with 2 participants
		wsSlice := runParticipantStubs(ns, perSession)
		for index, ws := range wsSlice {
			ws.landWith(fmt.Sprintf("fingerprint%v", index)).agree()
		}

		// assert session has started
		for _, ws := range wsSlice {
			assert.True(t, retryUntil(longerDuration, func() bool {
				return ws.hasReceivedKind("Starting")
			}), "participant should receive Starting with oTree starting link")
		}

		// first participant reconnects (same fingerprint)
		time.Sleep((sessionDurationTest + 1) * models.SessionDurationUnit)
		wsSlice[0].Close()
		ws := runParticipantStub(ns)
		ws.landWith("fingerprint0").agree()

		assert.False(t, retryUntil(longerDuration, func() bool {
			ok := ws.hasReceived(Message{"Disconnect", "LandingFailed"})
			return ok
		}))
	})
}
