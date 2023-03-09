package live

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunner_Integration(t *testing.T) {
	t.Run("runner is shared per campaign", func(t *testing.T) {
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
			return ws1.lastWrite() == "PoolSize:2"
		}), "participant should receive PoolSize:2")
		assert.True(t, retryUntil(shortDuration, func() bool {
			return ws2.lastWrite() == "PoolSize:2"
		}), "participant should receive PoolSize:2")
	})
}
