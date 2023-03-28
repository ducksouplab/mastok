package live

import (
	"errors"

	"github.com/ducksouplab/mastok/helpers"
)

type wsStub struct {
	doneCh    chan struct{}
	toRead    chan Message
	writtenTo chan Message
	// internal
	writes []Message
}

func newWSStub() *wsStub {
	ws := &wsStub{
		doneCh:    make(chan struct{}),
		toRead:    make(chan Message, 256),
		writtenTo: make(chan Message, 256),
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
		case msg := <-ws.toRead:
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
		case ws.writtenTo <- m.(Message):
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

// internals
// ----

// to the other side of the websocket, we may push (for future ReadJSON)
// or pull (what has been WriteJSON)
func (ws *wsStub) push(m Message) {
	ws.toRead <- m
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
	for w := range ws.writtenTo {
		ws.writes = append(ws.writes, w)
	}
}
