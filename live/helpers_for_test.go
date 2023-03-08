package live

import (
	"errors"

	th "github.com/ducksouplab/mastok/test_helpers"
)

type wsStub struct {
	closed bool
}

func (ws *wsStub) ReadJSON(m any) error {
	if ws.closed {
		return errors.New("stub closed")
	}
	return nil
}

func (ws *wsStub) WriteJSON(m any) error {
	if ws.closed {
		return errors.New("stub closed")
	}
	return nil
}

// closing stops reading, which in turn unregister client from runner
// which in turn deletes runner from store (if was the last registered client) and stops it loop
func (ws *wsStub) Close() error {
	ws.closed = true
	return nil
}

func newWSStub() *wsStub {
	return &wsStub{closed: false}
}

func init() {
	// CAUTION: currently DB is not reinitialized after each test, but at a package level
	th.ReinitTestDB()
}

func getRunnerStoreSize() int {
	rs.Lock()
	defer rs.Unlock()

	return len(rs.index)
}
