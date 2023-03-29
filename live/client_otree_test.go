package live

import (
	"strings"
	"testing"

	"github.com/ducksouplab/mastok/models"
	th "github.com/ducksouplab/mastok/test_helpers"
	"github.com/stretchr/testify/assert"
)

func TestClient_Otree_Integration(t *testing.T) {

	t.Run("creates oTree session and sends relevant SessionStart to participants and supervisors", func(t *testing.T) {
		ns := "fxt_otree_to_be_launched"
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

	t.Run("pool and pending are updated after StartSession", func(t *testing.T) {
		ns := "fxt_otree_groups_to_be_launched"
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

		wsMales := runParticipantStubs(ns, 5)
		wsFemales := runParticipantStubs(ns, 3)

		for _, ws := range wsMales {
			ws.land().agree().choose("Male")
		}

		for _, ws := range wsFemales {
			ws.land().agree().choose("Female")
		}

		assert.True(t, retryUntil(longerDuration, func() bool {
			_, ok := wsSup.hasReceivedKind("SessionStart")
			return ok
		}))

		wsSup.ClearTillMessage(Message{"SessionStart", "*"})
		assert.True(t, retryUntil(shortDuration, func() bool {
			ok := wsSup.hasReceived(Message{"PoolSize", "3/4"})
			return ok
		}))
	})

}
