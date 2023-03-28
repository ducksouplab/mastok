package live

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunner_Integration(t *testing.T) {
	t.Run("runner is shared per campaign", func(t *testing.T) {
		ns := "fxt_live_ns1"
		slug := ns + "_slug"
		defer tearDown(ns)

		// first client
		ws1 := newWSStub()
		p1 := RunParticipant(ws1, slug)
		runner := p1.runner

		// second client
		ws2 := newWSStub()
		p2 := RunParticipant(ws2, slug)
		// runner is shared
		assert.Same(t, runner, p2.runner, "participants runner should be the same")

		// participants agree
		ws1.land().agree()
		ws2.land().agree()
		// clients write PoolSize
		assert.True(t, retryUntil(shortDuration, func() bool {
			return ws1.isLastWrite(Message{"PoolSize", "2/4"})
		}), "participant should receive PoolSize:2/4")
		assert.True(t, retryUntil(shortDuration, func() bool {
			return ws2.isLastWrite(Message{"PoolSize", "2/4"})
		}), "participant should receive PoolSize:2/4")
	})

	t.Run("cleans up runner when closed", func(t *testing.T) {
		ns := "fxt_live_ns1"
		slug := ns + "_slug"
		// no teardown since we are actually testing the effects of quitting (ws.Close())

		// two clients
		ws1 := newWSStub()
		ws2 := newWSStub()
		RunSupervisor(ws1, ns)
		RunParticipant(ws2, slug)
		// both clients are here
		sharedRunner, ok := hasRunner(ns)
		assert.Equal(t, true, ok)

		// one quits
		ws1.Close()
		_, ok = hasRunner(ns)
		assert.Equal(t, true, ok)

		// the other quits
		ws2.Close()
		<-sharedRunner.isDone()
		_, ok = hasRunner(ns)
		assert.Equal(t, false, ok)
	})
}
