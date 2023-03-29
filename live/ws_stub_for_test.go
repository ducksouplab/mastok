package live

import (
	"errors"
	"log"

	"github.com/ducksouplab/mastok/helpers"
	"github.com/ducksouplab/mastok/models"
)

type wsStub struct {
	// for logging
	label string
	// channels
	toReadCh    chan Message
	writtenToCh chan Message
	clearAllCh  chan struct{}
	clearTillCh chan Message
	// closing
	done   bool
	doneCh chan struct{}
	// internal
	writes []Message
}

func runSupervisorStub(ns string) (ws *wsStub, campaign *models.Campaign) {
	ws = newWSStub(ns + "#supervisor")
	sup := RunSupervisor(ws, ns)
	campaign = sup.runner.campaign
	return
}

func runParticipantStub(ns string) (ws *wsStub) {
	ws = newWSStub(ns + "#participant")
	RunParticipant(ws, ns+"_slug")
	return
}

func runParticipantStubs(ns string, size int) (wsSlice []*wsStub) {
	for i := 0; i < size; i++ {
		ws := newWSStub(ns + "#participant")
		RunParticipant(ws, ns+"_slug")
		wsSlice = append(wsSlice, ws)
	}
	return
}

func newWSStub(l string) *wsStub {
	ws := &wsStub{
		label:       l,
		toReadCh:    make(chan Message, 256),
		writtenToCh: make(chan Message, 256),
		clearAllCh:  make(chan struct{}),
		clearTillCh: make(chan Message),
		doneCh:      make(chan struct{}),
	}
	go ws.loop()
	return ws
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
			return errors.New("[stub] done while ReadJSON for " + ws.label)
		}
	}
}

func (ws *wsStub) WriteJSON(m any) error {
	for {
		select {
		case ws.writtenToCh <- m.(Message):
			return nil
		case <-ws.doneCh:
			return errors.New("[stub] done while WriteJSON for " + ws.label)
		}
	}
}

func (ws *wsStub) ClearAllMessages() {
	ws.clearAllCh <- struct{}{}
}

func (ws *wsStub) ClearTillMessage(m Message) {
	ws.clearTillCh <- m
}

// closing stops reading, which in turn unregister client from runner
// which in turn deletes runner from store (if was the last registered client) and stops it loop
func (ws *wsStub) Close() error {
	if !ws.done {
		ws.done = true
		log.Printf("[stub] Close called for " + ws.label)
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
		case <-ws.clearAllCh:
			ws.writes = []Message{}
		case till := <-ws.clearTillCh:
			discard := true
			newWrites := []Message{}
			for _, m := range ws.writes {
				if !discard {
					newWrites = append(newWrites, m)
				}
				if till.Kind == m.Kind && (till.Payload == "*" || till.Payload == m.Payload) {
					discard = false
				}
			}
			ws.writes = newWrites
		}
	}
}
