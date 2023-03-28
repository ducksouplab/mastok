package live

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Participant_Grouping_Integration(t *testing.T) {

	t.Run("with grouping, PoolSize is sent after Choose", func(t *testing.T) {
		ns := "fxt_live_ns14_grouping"
		slug := ns + "_slug"
		defer tearDown(ns)

		ws1 := newWSStub()
		p1 := RunParticipant(ws1, slug)
		ws2 := newWSStub()
		RunParticipant(ws2, slug)

		// the fixture data is what we expected
		assert.Contains(t, p1.runner.campaign.Grouping, "Male")

		assert.False(t, retryUntil(shortDuration, func() bool {
			_, ok := ws1.hasReceivedKind("PoolSize")
			return ok
		}))

		ws1.land()

		assert.False(t, retryUntil(shortDuration, func() bool {
			_, ok := ws1.hasReceivedKind("PoolSize")
			return ok
		}))

		ws1.agree()

		assert.False(t, retryUntil(shortDuration, func() bool {
			_, ok := ws1.hasReceivedKind("PoolSize")
			return ok
		}))

		ws1.choose("Male")

		assert.True(t, retryUntil(shortDuration, func() bool {
			_, ok := ws1.hasReceivedKind("PoolSize")
			return ok
		}))

		ws2.land().agree()
		assert.False(t, retryUntil(shortDuration, func() bool {
			_, ok := ws2.hasReceivedKind("PoolSize")
			return ok
		}))
	})

}
