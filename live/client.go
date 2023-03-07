package live

import (
	"log"

	"github.com/ducksouplab/mastok/types"
	"github.com/gorilla/websocket"
)

// client for campaign runner
type client struct {
	supervisor bool
	ws         *wsConn
	runner     *runner
	signal     chan types.Message
}

func (c *client) readLoop() {
	defer func() {
		c.runner.unregisterCh <- c
		c.ws.Close()
	}()

	for {
		m, err := c.ws.read()

		if err != nil {
			log.Println(err)
			return
		} else if m.Kind == "State" {
			c.runner.stateCh <- m.Payload
		} else if m.Kind == "Join" {
			c.runner.joinCh <- m.Payload
		}
	}
}

func (c *client) writeLoop() {
	for {
		state := <-c.signal

		if state.Kind == "State" {
			c.ws.write(state)
		}
	}
}

func runClient(s bool, conn *websocket.Conn, namespace string) {
	r, err := getRunner(namespace)
	if err != nil {
		conn.Close()
		return
	}
	c := &client{
		supervisor: s,
		ws:         newWsConn(conn),
		runner:     r,
		signal:     make(chan types.Message),
	}
	log.Println("[supervisor] running for: " + namespace)

	c.runner.registerCh <- c
	go c.readLoop()
	go c.writeLoop()
}

func RunSupervisor(conn *websocket.Conn, namespace string) {
	runClient(true, conn, namespace)
}

func RunParticipant(conn *websocket.Conn, namespace string) {
	runClient(false, conn, namespace)
}
