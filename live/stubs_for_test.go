package live

import (
	"errors"

	"github.com/ducksouplab/mastok/helpers"
)

type wsStub struct {
	toReadCh    chan Message
	writtenToCh chan Message
	clearCh     chan struct{}
	// closing
	done   bool
	doneCh chan struct{}
	// internal
	writes []Message
}

func newWSStub() *wsStub {
	ws := &wsStub{
		toReadCh:    make(chan Message, 256),
		writtenToCh: make(chan Message, 256),
		clearCh:     make(chan struct{}),
		doneCh:      make(chan struct{}),
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

// API for the supervisor and participant (client.go)
// ----
func (ws *wsStub) ReadJSON(m any) error {
	for {
		select {
		case msg := <-ws.toReadCh:
			pointer := m.(*Message)
			*pointer = msg
			return nil
		case <-ws.doneCh:
			return errors.New("ws stub closed")
		}
	}
}

func (ws *wsStub) WriteJSON(m any) error {
	for {
		select {
		case ws.writtenToCh <- m.(Message):
			return nil
		case <-ws.doneCh:
			return errors.New("ws stub closed")
		}
	}
}

func (ws *wsStub) Clear() {
	ws.clearCh <- struct{}{}
}

// closing stops reading, which in turn unregister client from runner
// which in turn deletes runner from store (if was the last registered client) and stops it loop
func (ws *wsStub) Close() error {
	if !ws.done {
		ws.done = true
		close(ws.doneCh)
	}
	return nil
}

// internals
// ----

// to the other side of the websocket, we may push (for future ReadJSON)
// or pull (what has been WriteJSON)
func (ws *wsStub) push(m Message) {
	ws.toReadCh <- m
}

// write helpers
func (ws *wsStub) landWith(fingerprint string) *wsStub {
	ws.push(Message{"Land", fingerprint})
	return ws
}

func (ws *wsStub) land() *wsStub {
	ws.push(Message{"Land", helpers.RandomHexString(64)})
	return ws
}

func (ws *wsStub) agree() *wsStub {
	ws.push(Message{"Agree", ""})
	return ws
}

func (ws *wsStub) choose(groupLabel string) *wsStub {
	ws.push(Message{"Choose", groupLabel})
	return ws
}

func (ws *wsStub) isLastWriteKind(kind string) bool {
	length := len(ws.writes)
	if length == 0 {
		return false
	}
	last := ws.writes[length-1]
	return last.Kind == kind
}

func (ws *wsStub) isLastWrite(m Message) bool {
	length := len(ws.writes)
	if length == 0 {
		return false
	}
	last := ws.writes[length-1]
	return last.Kind == m.Kind && last.Payload == m.Payload
}

func (ws *wsStub) hasReceived(test Message) bool {
	for _, write := range ws.writes {
		if write.Kind == test.Kind && write.Payload == test.Payload {
			return true
		}
	}
	return false
}

func (ws *wsStub) hasReceivedKind(kind string) (found Message, ok bool) {
	for _, write := range ws.writes {
		if write.Kind == kind {
			return write, true
		}
	}
	return Message{}, false
}

func (ws *wsStub) loop() {
	for {
		select {
		case w := <-ws.writtenToCh:
			ws.writes = append(ws.writes, w)
		case <-ws.clearCh:
			ws.writes = []Message{}
		}
	}
}
