package live

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Grouping_Integration(t *testing.T) {

	t.Run("with grouping, JoiningSize is sent after Connect", func(t *testing.T) {
		ns := "fxt_grp"
		defer tearDown(ns)

		wsSup, campaign := runSupervisorStub(ns)

		ws1 := runParticipantStub(ns)
		ws2 := runParticipantStub(ns)

		// the fixture data is what we expected
		assert.Contains(t, campaign.Grouping, "Male")

		assert.False(t, retryUntil(shortDuration, func() bool {
			return ws1.hasReceivedKind("JoiningSize")
		}))

		ws1.land()

		assert.False(t, retryUntil(shortDuration, func() bool {
			return ws1.hasReceivedKind("JoiningSize")
		}))

		ws1.agree()

		assert.False(t, retryUntil(shortDuration, func() bool {
			return ws1.hasReceivedKind("JoiningSize")
		}))

		ws1.connectWithGroup("Male")

		assert.True(t, retryUntil(shortDuration, func() bool {
			return ws1.hasReceivedKind("JoiningSize")
		}))
		assert.True(t, retryUntil(shortDuration, func() bool {
			return wsSup.hasReceivedKind("JoiningSize")
		}))

		ws2.land().agree()
		assert.False(t, retryUntil(shortDuration, func() bool {
			return ws2.hasReceivedKind("JoiningSize")
		}))
	})
}
