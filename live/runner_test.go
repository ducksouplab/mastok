package live

import (
	"strings"
	"testing"

	"github.com/ducksouplab/mastok/env"
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
		perSession := 4
		defer tearDown(ns)

		// the fixture data is what we expected
		campaign, _ := models.FindCampaignByNamespace(ns)
		assert.Equal(t, perSession, campaign.PerSession)
		assert.Equal(t, "Running", campaign.State)

		// fills the pool
		th.InterceptOtreePostSession()
		th.InterceptOtreeGetSession("/api/sessions/")
		defer th.InterceptOff()
		wsSlice := makeWSStubs(perSession)
		for _, ws := range wsSlice {
			RunParticipant(ws, ns)
		}

		assert.True(t, retryUntil(longerDuration, func() bool {
			found, ok := wsSlice[0].hasReceivedPrefix("SessionStart:")
			if ok {
				url := strings.TrimPrefix(found, "SessionStart:")
				//http://localhost:8180/InitializeParticipant/brutjmj7
				return strings.HasPrefix(url, env.OTreeURL+"/InitializeParticipant/")
			}
			return false
		}), "participant should receive SessionState with oTree URL")
	})
}
