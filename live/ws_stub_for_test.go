package live

import (
	"errors"
	"strings"
)

type wsStub struct {
	doneCh    chan struct{}
	toRead    chan string
	writtenTo chan string
	// internal
	writes []string
}

// API for the supervisor and participant clients
func (ws *wsStub) ReadJSON(m any) error {
	for {
		select {
		case read := <-ws.toRead:
			p := m.(*string)
			*p = read
			return nil
		case <-ws.doneCh:
			return errors.New("ws stub closed")
		}
	}
}

func (ws *wsStub) WriteJSON(m any) error {
	ms := m.(string)
	for {
		select {
		case ws.writtenTo <- ms:
			return nil
		case <-ws.doneCh:
			return errors.New("ws stub closed")
		}
	}
}

// closing stops reading, which in turn unregister client from runner
// which in turn deletes runner from store (if was the last registered client) and stops it loop
func (ws *wsStub) Close() error {
	close(ws.doneCh)
	return nil
}

// to the other side of the websocket, we may push (for future ReadJSON)
// or pull (what has been WriteJSON)
func (ws *wsStub) push(m string) {
	ws.toRead <- m
}

func (ws *wsStub) lastWrite() string {
	if length := len(ws.writes); length == 0 {
		return ""
	} else {
		return ws.writes[length-1]
	}
}

func (ws *wsStub) hasReceived(test string) bool {
	for _, write := range ws.writes {
		if write == test {
			return true
		}
	}
	return false
}

func (ws *wsStub) hasReceivedPrefix(prefix string) (found string, ok bool) {
	for _, write := range ws.writes {
		if strings.HasPrefix(write, prefix) {
			return write, true
		}
	}
	return "", false
}

func (ws *wsStub) loop() {
	for w := range ws.writtenTo {
		ws.writes = append(ws.writes, w)
	}
}

func newWSStub() *wsStub {
	ws := &wsStub{
		doneCh:    make(chan struct{}),
		toRead:    make(chan string, 256),
		writtenTo: make(chan string, 256),
	}
	go ws.loop()
	return ws
}

func makeWSStubs(size int) []*wsStub {
	out := make([]*wsStub, size)
	for i := 0; i < size; i++ {
		out[i] = newWSStub()
	}
	return out
}
