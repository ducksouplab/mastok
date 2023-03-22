package live

import (
	"strings"
	"testing"

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
		campaign, _ := models.FindCampaignByNamespace(ns)
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
		campaign, _ := models.FindCampaignByNamespace(ns)
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
			c, _ := models.FindCampaignByNamespace(ns)
			return c.State == models.Completed && c.StartedSessions == 4
		}), "campaign should be Completed")

		// outer state: new participant can't connect
		addWs := newWSStub()
		RunParticipant(addWs, slug)
		assert.True(t, retryUntil(shortDuration, func() bool {
			return addWs.isLastWriteKind("Disconnect")
		}), "participant should receive Disconnect")
	})

	// t.Run("sends redirect if same participant reconnects while session is live", func(t *testing.T) {
	// 	ns := "fxt_live_ns9_rejoin"
	// 	slug := ns + "_slug"
	// 	perSession := 2
	// 	defer tearDown(ns)

	// 	// the fixture data is what we expected
	// 	campaign, _ := models.FindCampaignByNamespace(ns)
	// 	assert.Equal(t, perSession, campaign.PerSession)
	// 	assert.Equal(t, 0, campaign.StartedSessions)
	// 	assert.Equal(t, "Running", campaign.State)

	// 	// fills the pool
	// 	th.InterceptOtreePostSession()
	// 	th.InterceptOtreeGetSession()
	// 	defer th.InterceptOff()

	// 	// complete pool with 2 participants
	// 	wsSlice := makeWSStubs(perSession)
	// 	for index, ws := range wsSlice {
	// 		RunParticipant(ws, slug)
	// 		ws.landWith(fmt.Sprintf("fingerprint%v", index)).join()
	// 	}

	// 	// first participant reconnects
	// 	time.Sleep(shortDuration)
	// 	wsSlice[0].Close()
	// 	ws := newWSStub()
	// 	RunParticipant(ws, slug)
	// 	ws.landWith("fingerprint0").join()

	// 	assert.True(t, retryUntil(longDuration, func() bool {
	// 		_, found := ws.hasReceivedKind("Redirect")
	// 		return found
	// 	}), "participant should receive Redirect")
	// })

	// t.Run("sends reject if same participant reconnects while session is not live", func(t *testing.T) {

	// })
}
