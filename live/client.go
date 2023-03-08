package live

import (
	"log"
	"strconv"

	"github.com/ducksouplab/mastok/types"
)

type wsConn interface {
	ReadJSON(any) error
	WriteJSON(any) error
	Close() error
}

// client for campaign runner
type client struct {
	supervisor bool
	ws         wsConn
	runner     *runner
	stateCh    chan string
	poolSizeCh chan int
}

func (c *client) writeMessage(kind, payload string) {
	if err := c.ws.WriteJSON(types.Message{Kind: kind, Payload: payload}); err != nil {
		log.Println(err)
	}
}

func (c *client) readLoop() {
	defer func() {
		c.runner.leavePoolCh <- c
		c.runner.unregisterCh <- c
	}()

	for {
		var m types.Message
		err := c.ws.ReadJSON(&m)

		if err != nil {
			log.Println("[ws] ReadJSON error:", err)
			return
		} else if m.Kind == "State" {
			c.runner.updateStateCh <- m.Payload
		} else if m.Kind == "Join" {
			c.runner.joinPoolCh <- c
		}
	}
}

// at most one writer to a connection since all writes happen in this goroutine
// like in https://github.com/gorilla/websocket/blob/master/examples/chat/client.go
func (c *client) writeLoop() {
	for {
		select {
		case state := <-c.stateCh:
			c.writeMessage("State", state)
		case size := <-c.poolSizeCh:
			c.writeMessage("PoolSize", strconv.Itoa(size))
		}
	}
}

func runClient(s bool, ws wsConn, namespace string) *client {
	r, err := getRunner(namespace)
	if err != nil {
		ws.Close()
		return nil
	}
	c := &client{
		supervisor: s,
		ws:         ws,
		runner:     r,
		stateCh:    make(chan string),
		poolSizeCh: make(chan int),
	}
	log.Println("[supervisor] running for: " + namespace)

	c.runner.registerCh <- c
	go c.readLoop()
	go c.writeLoop()
	return c
}

func RunSupervisor(ws wsConn, namespace string) *client {
	return runClient(true, ws, namespace)
}

func RunParticipant(ws wsConn, namespace string) *client {
	p := runClient(false, ws, namespace)
	p.runner.joinPoolCh <- p
	return p
}
