package live

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunner_Integration(t *testing.T) {
	t.Run("has runner", func(t *testing.T) {
		ns := "fixture_ns1"
		// first client
		ws1 := newWSStub()
		p1 := RunParticipant(ws1, ns)
		defer ws1.Close()
		runner := p1.runner
		// second client
		ws2 := newWSStub()
		p2 := RunParticipant(ws2, ns)
		defer ws2.Close()
		assert.Same(t, runner, p2.runner, "participants runner should be the same")

		time.Sleep(10 * time.Millisecond)
		assert.Equal(t, 2, runner.poolSize)

	})

}
