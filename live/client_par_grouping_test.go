package live

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Grouping_Integration(t *testing.T) {

	t.Run("with grouping, PoolSize is sent after Choose", func(t *testing.T) {
		ns := "fxt_grp"
		defer tearDown(ns)

		wsSup, campaign := runSupervisorStub(ns)

		ws1 := runParticipantStub(ns)
		ws2 := runParticipantStub(ns)

		// the fixture data is what we expected
		assert.Contains(t, campaign.Grouping, "Male")

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
		assert.True(t, retryUntil(shortDuration, func() bool {
			_, ok := wsSup.hasReceivedKind("PoolSize")
			return ok
		}))

		ws2.land().agree()
		assert.False(t, retryUntil(shortDuration, func() bool {
			_, ok := ws2.hasReceivedKind("PoolSize")
			return ok
		}))
	})
}
