package live

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_Integration(t *testing.T) {
	t.Run("has runner", func(t *testing.T) {
		ns := "fixture_ns1"
		ws := newWSStub()
		RunSupervisor(ws, ns)
		defer ws.Close()

		_, ok := hasRunner(ns)
		t.Log(rs.index)
		assert.Equal(t, true, ok)
	})

	t.Run("cleans up runner when closed", func(t *testing.T) {
		ns := "fixture_ns1"
		// two clients
		ws1 := newWSStub()
		ws2 := newWSStub()
		RunSupervisor(ws1, ns)
		RunParticipant(ws2, ns)
		// both clients are here
		_, ok := hasRunner(ns)
		assert.Equal(t, true, ok)
		// one quits
		ws1.Close()
		time.Sleep(100 * time.Millisecond)
		_, ok = hasRunner(ns)
		assert.Equal(t, true, ok)
		// the other quits
		ws2.Close()
		time.Sleep(100 * time.Millisecond)
		_, ok = hasRunner(ns)
		assert.Equal(t, false, ok)
	})
}
