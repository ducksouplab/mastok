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
		ns := "fxt_live_ns5_launched"
		slug := ns + "_slug"
		perSession := 4
		defer tearDown(ns)

		// the fixture data is what we expected
		campaign, _ := models.GetCampaignByNamespace(ns)
		assert.Equal(t, perSession, campaign.PerSession)
		assert.Equal(t, 0, campaign.StartedSessions)
		assert.Equal(t, "Running", campaign.State)

		// fills the pool
		th.InterceptOtreePostSession()
		th.InterceptOtreeGetSession()
		defer th.InterceptOff()
		// 1 supervisor
		wsSupSlice := makeWSStubs(2)
		for _, wsSup := range wsSupSlice {
			RunSupervisor(wsSup, ns)
		}
		// 4 participants
		wsSlice := makeWSStubs(perSession)
		for _, ws := range wsSlice {
			RunParticipant(ws, slug)
			ws.land().join()
		}

		assert.True(t, retryUntil(longerDuration, func() bool {
			found, ok := wsSupSlice[0].hasReceivedKind("SessionStart")
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
		ns := "fxt_live_ns7_almost_completed"
		slug := ns + "_slug"
		perSession := 4
		defer tearDown(ns)

		// the fixture data is what we expected
		campaign, _ := models.GetCampaignByNamespace(ns)
		assert.Equal(t, perSession, campaign.PerSession)
		assert.Equal(t, 3, campaign.StartedSessions)
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
		assert.True(t, retryUntil(longDuration, func() bool {
			c, _ := models.GetCampaignByNamespace(ns)
			return c.State == models.Completed && c.StartedSessions == 4
		}), "campaign should be Completed")

		// outer state: new participant can't connect
		addWs := newWSStub()
		RunParticipant(addWs, slug)
		// no need to land/join, Completed State will kick participant first thing
		assert.True(t, retryUntil(shortDuration, func() bool {
			return addWs.isLastWriteKind("Disconnect")
		}), "participant should receive Disconnect")
	})

	t.Run("sends redirect if participant reconnects", func(t *testing.T) {
		ns := "fxt_live_ns10_redirect"
		slug := ns + "_slug"
		perSession := 2
		defer tearDown(ns)

		// the fixture data is what we expected
		campaign, _ := models.GetCampaignByNamespace(ns)
		assert.Equal(t, perSession, campaign.PerSession)
		assert.Equal(t, false, campaign.JoinOnce)
		assert.Equal(t, 0, campaign.StartedSessions)
		assert.Equal(t, "Running", campaign.State)

		// fills the pool
		th.InterceptOtreePostSession()
		th.InterceptOtreeGetSession()
		defer th.InterceptOff()

		// complete pool with 2 participants
		wsSlice := makeWSStubs(perSession)
		for index, ws := range wsSlice {
			RunParticipant(ws, slug)
			ws.landWith(fmt.Sprintf("fingerprint%v", index)).join()
		}

		// assert session has started
		for _, ws := range wsSlice {
			assert.True(t, retryUntil(longDuration, func() bool {
				_, ok := ws.hasReceivedKind("SessionStart")
				return ok
			}), "participant should receive SessionStart with oTree starting link")
		}

		// first participant reconnects
		wsSlice[0].Close()
		ws := newWSStub()
		RunParticipant(ws, slug)
		ws.landWith("fingerprint0").join()

		assert.True(t, retryUntil(longDuration, func() bool {
			_, found := ws.hasReceivedKind("Redirect")
			return found
		}), "participant should receive Redirect")
	})

	t.Run("sends redirect if participant reconnects even for JoinOnce campaign", func(t *testing.T) {
		ns := "fxt_live_ns11_redirect2"
		slug := ns + "_slug"
		perSession := 2
		defer tearDown(ns)

		// the fixture data is what we expected
		campaign, _ := models.GetCampaignByNamespace(ns)
		assert.Equal(t, perSession, campaign.PerSession)
		assert.Equal(t, true, campaign.JoinOnce)
		assert.Equal(t, 0, campaign.StartedSessions)
		assert.Equal(t, "Running", campaign.State)

		// fills the pool
		th.InterceptOtreePostSession()
		th.InterceptOtreeGetSession()
		defer th.InterceptOff()

		// complete pool with 2 participants
		wsSlice := makeWSStubs(perSession)
		for index, ws := range wsSlice {
			RunParticipant(ws, slug)
			ws.landWith(fmt.Sprintf("fingerprint%v", index)).join()
		}

		// assert session has started
		for _, ws := range wsSlice {
			assert.True(t, retryUntil(longDuration, func() bool {
				_, ok := ws.hasReceivedKind("SessionStart")
				return ok
			}), "participant should receive SessionStart with oTree starting link")
		}

		// first participant reconnects
		wsSlice[0].Close()
		ws := newWSStub()
		RunParticipant(ws, slug)
		ws.landWith("fingerprint0").join()

		assert.True(t, retryUntil(longDuration, func() bool {
			_, found := ws.hasReceivedKind("Redirect")
			return found
		}), "participant should receive Redirect")
	})

	t.Run("sends reject if participant reconnect to JoinOnce campaign after session ended", func(t *testing.T) {
		ns := "fxt_live_ns12_reject"
		slug := ns + "_slug"
		perSession := 2
		defer tearDown(ns)

		// the fixture data is what we expected
		campaign, _ := models.GetCampaignByNamespace(ns)
		assert.Equal(t, perSession, campaign.PerSession)
		assert.Equal(t, true, campaign.JoinOnce)
		assert.Equal(t, 0, campaign.StartedSessions)
		assert.Equal(t, "Running", campaign.State)

		// fills the pool
		th.InterceptOtreePostSession()
		th.InterceptOtreeGetSession()
		defer th.InterceptOff()

		// complete pool with 2 participants
		wsSlice := makeWSStubs(perSession)
		for index, ws := range wsSlice {
			RunParticipant(ws, slug)
			ws.landWith(fmt.Sprintf("fingerprint%v", index)).join()
		}

		// assert session has started
		for _, ws := range wsSlice {
			assert.True(t, retryUntil(longDuration, func() bool {
				_, ok := ws.hasReceivedKind("SessionStart")
				return ok
			}), "participant should receive SessionStart with oTree starting link")
		}

		// first participant reconnects
		time.Sleep((sessionDurationTest + 1) * models.SessionDurationUnit)
		wsSlice[0].Close()
		ws := newWSStub()
		RunParticipant(ws, slug)
		ws.landWith("fingerprint0").join()

		assert.True(t, retryUntil(longDuration, func() bool {
			_, found := ws.hasReceivedKind("Reject")
			return found
		}))
	})

	t.Run("does not send reject if participant reconnect to multi-join campaign after session ended", func(t *testing.T) {
		ns := "fxt_live_ns13_noreject"
		slug := ns + "_slug"
		perSession := 2
		defer tearDown(ns)

		// the fixture data is what we expected
		campaign, _ := models.GetCampaignByNamespace(ns)
		assert.Equal(t, perSession, campaign.PerSession)
		assert.Equal(t, false, campaign.JoinOnce)
		assert.Equal(t, 0, campaign.StartedSessions)
		assert.Equal(t, "Running", campaign.State)

		// fills the pool
		th.InterceptOtreePostSession()
		th.InterceptOtreeGetSession()
		defer th.InterceptOff()

		// complete pool with 2 participants
		wsSlice := makeWSStubs(perSession)
		for index, ws := range wsSlice {
			RunParticipant(ws, slug)
			ws.landWith(fmt.Sprintf("fingerprint%v", index)).join()
		}

		// assert session has started
		for _, ws := range wsSlice {
			assert.True(t, retryUntil(longDuration, func() bool {
				_, ok := ws.hasReceivedKind("SessionStart")
				return ok
			}), "participant should receive SessionStart with oTree starting link")
		}

		// first participant reconnects
		time.Sleep((sessionDurationTest + 1) * models.SessionDurationUnit)
		wsSlice[0].Close()
		ws := newWSStub()
		RunParticipant(ws, slug)
		ws.landWith("fingerprint0").join()

		assert.False(t, retryUntil(longDuration, func() bool {
			_, found := ws.hasReceivedKind("Reject")
			return found
		}))
	})
}
