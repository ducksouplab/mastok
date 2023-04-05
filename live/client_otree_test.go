package live

import (
	"strings"
	"testing"
	"time"

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
			found, ok := wsSup1.firstOfKind("SessionStart")
			if ok {
				session := found.Payload.(models.Session)
				return strings.Contains(session.AdminUrl, "/SessionStartLinks/")
			}
			return false
		}), "supervisor should receive SessionStart with oTree admin URL and oTree id like mk:namespace:#")

		urlsMap := map[string]bool{}
		for _, ws := range wsSlice {
			assert.True(t, retryUntil(longDuration, func() bool {
				found, ok := ws.firstOfKind("SessionStart")
				if ok {
					url := found.Payload.(string)
					urlsMap[url] = true
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
			ws.land().agree().connectWithGroup("Male")
		}
		for _, ws := range wsFemales {
			ws.land().agree().connectWithGroup("Female")
		}

		assert.True(t, retryUntil(longerDuration, func() bool {
			return wsSup.hasReceivedKind("SessionStart")
		}))

		wsSup.ClearTillMessage(Message{"SessionStart", "*"})
		assert.True(t, retryUntil(shortDuration, func() bool {
			return wsSup.hasReceived(Message{"PoolSize", "3/4"})
		}))
	})

	t.Run("two StartSession when pending is big enough", func(t *testing.T) {
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
		wsFemales := runParticipantStubs(ns, 1)
		for _, ws := range wsMales {
			ws.land().agree().connectWithGroup("Male")
		}
		for _, ws := range wsFemales {
			ws.land().agree().connectWithGroup("Female")
		}

		// missing a Female
		assert.False(t, retryUntil(longerDuration, func() bool {
			return wsSup.hasReceivedKind("SessionStart")
		}))

		wsFemalesToo := runParticipantStubs(ns, 3)
		for _, ws := range wsFemalesToo {
			ws.land().agree().connectWithGroup("Female")
		}

		assert.True(t, retryUntil(longerDuration, func() bool {
			return wsSup.hasReceivedKind("SessionStart")
		}))

		wsSup.ClearTillMessage(Message{"SessionStart", "*"})
		assert.True(t, retryUntil(3*longerDuration, func() bool {
			return wsSup.hasReceivedKind("SessionStart")
		}))
	})

	t.Run("StartSession waits if concurrent session limit is reached", func(t *testing.T) {
		ns := "fxt_otree_concurrent"
		defer tearDown(ns)

		wsSup, campaign := runSupervisorStub(ns)

		// // the fixture data is what we expected
		assert.Equal(t, 2, campaign.PerSession)
		assert.Equal(t, 2, campaign.ConcurrentSessions)
		assert.Equal(t, campaign.Grouping, "")

		// // oTree interception is needed when SessionStart
		th.InterceptOtreePostSession()
		th.InterceptOtreeGetSession()
		defer th.InterceptOff()

		wsParticipants := runParticipantStubs(ns, 8)
		for _, ws := range wsParticipants {
			ws.land().agree()
		}

		// caution: timing is a bit too empiric in this test
		time.Sleep(longDuration)
		sessionStartCount := wsSup.countKind("SessionStart")
		assert.Equal(t, 2, sessionStartCount)

		time.Sleep((sessionDurationTest + 1) * models.SessionDurationUnit)
		sessionStartCount = wsSup.countKind("SessionStart")
		assert.True(t, sessionStartCount >= 3)
	})

}
