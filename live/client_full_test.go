package live

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ducksouplab/mastok/models"
	th "github.com/ducksouplab/mastok/test_helpers"
	"github.com/stretchr/testify/assert"
)

func TestClientFull_Integration(t *testing.T) {

	t.Run("creates oTree session and sends relevant SessionStart to participants and supervisors", func(t *testing.T) {
		ns := "fxt_par_launched"
		perSession := 4
		defer tearDown(ns)

		// 2 supervisors
		wsSup1, campaign := runSupervisorStub(ns)
		runSupervisorStub(ns)

		// the fixture data is what we expected
		assert.Equal(t, perSession, campaign.PerSession)
		assert.Equal(t, 0, campaign.StartedSessions)
		assert.Equal(t, "Running", campaign.State)

		// fills the room
		th.InterceptOtreePostSession()
		th.InterceptOtreeGetSession()
		defer th.InterceptOff()

		// 4 participants
		wsSlice := runParticipantStubs(ns, perSession)
		for _, ws := range wsSlice {
			ws.land().agree()
		}

		assert.True(t, retryUntil(longerDuration, func() bool {
			found, ok := wsSup1.hasReceivedKind("SessionStart")
			if ok {
				session := found.Payload.(models.Session)
				//http://localhost:8180/SessionStartLinks/t1wlmb4v
				return strings.Contains(session.AdminUrl, "/SessionStartLinks/")
			}
			return false
		}), "supervisor should receive SessionStart with oTree admin URL and oTree id like mk:namespace:#")

		urlsMap := map[string]bool{}
		for _, ws := range wsSlice {
			assert.True(t, retryUntil(longDuration, func() bool {
				found, ok := ws.hasReceivedKind("SessionStart")
				if ok {
					url := found.Payload.(string)
					urlsMap[url] = true
					//http://localhost:8180/InitializeParticipant/brutjmj7
					return strings.Contains(url, "/InitializeParticipant/")
				}
				return false
			}), "participant should receive SessionStart with oTree starting link")
		}
		assert.Equal(t, len(wsSlice), len(urlsMap), "participants should received different oTree starting links")
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
				_, ok := ws.hasReceivedKind("SessionStart")
				return ok
			}), "participant should receive SessionStart with oTree starting link")
		}

		// first participant reconnects (same fingerprint)
		wsSlice[0].Close()
		ws := runParticipantStub(ns)
		ws.landWith("fingerprint0").agree()

		assert.True(t, retryUntil(longDuration, func() bool {
			_, found := ws.hasReceivedKind("Redirect")
			return found
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
				_, ok := ws.hasReceivedKind("SessionStart")
				return ok
			}), "participant should receive SessionStart with oTree starting link")
		}

		// first participant reconnects (same fingerprint)
		wsSlice[0].Close()
		ws := runParticipantStub(ns)
		ws.landWith("fingerprint0").agree()

		assert.True(t, retryUntil(longDuration, func() bool {
			_, found := ws.hasReceivedKind("Redirect")
			return found
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
				_, ok := ws.hasReceivedKind("SessionStart")
				return ok
			}), "participant should receive SessionStart with oTree starting link")
		}

		// first participant reconnects (same fingerprint)
		time.Sleep((sessionDurationTest + 1) * models.SessionDurationUnit)
		ws := runParticipantStub(ns)
		ws.landWith("fingerprint0").agree()

		assert.True(t, retryUntil(longDuration, func() bool {
			_, found := ws.hasReceivedKind("Reject")
			return found
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
				_, ok := ws.hasReceivedKind("SessionStart")
				return ok
			}), "participant should receive SessionStart with oTree starting link")
		}

		// first participant reconnects (same fingerprint)
		time.Sleep((sessionDurationTest + 1) * models.SessionDurationUnit)
		wsSlice[0].Close()
		ws := runParticipantStub(ns)
		ws.landWith("fingerprint0").agree()

		assert.False(t, retryUntil(longerDuration, func() bool {
			_, found := ws.hasReceivedKind("Reject")
			return found
		}))
	})
}
