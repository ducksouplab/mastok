package live

import (
	"encoding/json"

	"strings"
	"testing"

	"github.com/ducksouplab/mastok/models"
	th "github.com/ducksouplab/mastok/test_helpers"
	"github.com/stretchr/testify/assert"
)

func TestRunner_Integration(t *testing.T) {
	t.Run("is shared per campaign", func(t *testing.T) {
		ns := "fixture_ns1"
		defer tearDown(ns)

		// first client
		ws1 := newWSStub()
		p1 := RunParticipant(ws1, ns)
		runner := p1.runner

		// second client
		ws2 := newWSStub()
		p2 := RunParticipant(ws2, ns)
		// runner is shared
		assert.Same(t, runner, p2.runner, "participants runner should be the same")
		// clients write PoolSize
		assert.True(t, retryUntil(shortDuration, func() bool {
			return ws1.lastWrite() == "PoolSize:2/4"
		}), "participant should receive PoolSize:2/4")
		assert.True(t, retryUntil(shortDuration, func() bool {
			return ws2.lastWrite() == "PoolSize:2/4"
		}), "participant should receive PoolSize:2/4")
	})

	t.Run("creates oTree session", func(t *testing.T) {
		ns := "fixture_ns5_launched"
		sessionOtreeId := "mk:" + ns + ":1"
		perSession := 4
		defer tearDown(ns)

		// the fixture data is what we expected
		campaign, _ := models.FindCampaignByNamespace(ns)
		assert.Equal(t, perSession, campaign.PerSession)
		assert.Equal(t, 0, campaign.StartedSessions)
		assert.Equal(t, "Running", campaign.State)

		// fills the pool
		th.InterceptOtreePostSession()
		th.InterceptOtreeGetSession(sessionOtreeId)
		defer th.InterceptOff()
		// 1 supervisor
		wsSup := newWSStub()
		RunSupervisor(wsSup, ns)
		// 4 participants
		wsSlice := makeWSStubs(perSession)
		for _, ws := range wsSlice {
			RunParticipant(ws, ns)
		}

		assert.True(t, retryUntil(longerDuration, func() bool {
			found, ok := wsSlice[0].hasReceivedPrefix("SessionStart:")
			if ok {
				url := strings.TrimPrefix(found, "SessionStart:")
				//http://localhost:8180/InitializeParticipant/brutjmj7
				return strings.Contains(url, "/InitializeParticipant/")
			}
			return false
		}), "participant should receive SessionState with oTree URL")

		assert.True(t, retryUntil(shortDuration, func() bool {
			found, ok := wsSup.hasReceivedPrefix("SessionStart:")
			if ok {
				sessionMsh := strings.TrimPrefix(found, "SessionStart:")
				s := models.Session{}

				if err := json.Unmarshal([]byte(sessionMsh), &s); err != nil {
					t.Error("deserialize failed", sessionMsh)
				}
				//http://localhost:8180/SessionStartLinks/t1wlmb4v
				return strings.Contains(s.AdminUrl, "/SessionStartLinks/") && s.OtreeId == sessionOtreeId
			}
			return false
		}), "supervisor should receive SessionState with oTree admin URL and oTree id like mk:namespace:#")
	})

	t.Run("turns Campaign to completed after last SessionStart", func(t *testing.T) {
		ns := "fixture_ns7_almost_completed"
		perSession := 4
		defer tearDown(ns)

		// the fixture data is what we expected
		campaign, _ := models.FindCampaignByNamespace(ns)
		assert.Equal(t, perSession, campaign.PerSession)
		assert.Equal(t, 3, campaign.StartedSessions)
		assert.Equal(t, "Running", campaign.State)

		// fills the pool
		th.InterceptOtreePostSession()
		th.InterceptOtreeGetSession("mk:" + ns + ":4")
		defer th.InterceptOff()
		wsSlice := makeWSStubs(perSession)
		for _, ws := range wsSlice {
			RunParticipant(ws, ns)
		}

		// assert inner state
		assert.True(t, retryUntil(longerDuration, func() bool {
			c, _ := models.FindCampaignByNamespace(ns)
			return c.State == models.Completed && c.StartedSessions == 4
		}), "campaign should be Completed")

		// outer state: new participant can't connect
		addWs := newWSStub()
		RunParticipant(addWs, ns)
		assert.True(t, retryUntil(shortDuration, func() bool {
			return addWs.lastWrite() == "Participant:Disconnect"
		}), "participant should receive Disconnect")
	})
}
