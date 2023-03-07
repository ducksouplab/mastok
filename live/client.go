package live

import (
	"log"

	"github.com/ducksouplab/mastok/types"
	"github.com/gorilla/websocket"
)

// client for campaign runner
type client struct {
	supervisor bool
	ws         *websocket.Conn
	runner     *runner
	signal     chan types.Message
}

func (c *client) readLoop() {
	defer func() {
		c.runner.unregisterCh <- c
		c.ws.Close()
	}()

	for {
		var m types.Message
		err := c.ws.ReadJSON(&m)

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
			if err := c.ws.WriteJSON(state); err != nil {
				log.Println(err)
			}
		}
	}
}

func runClient(s bool, ws *websocket.Conn, namespace string) {
	r, err := getRunner(namespace)
	if err != nil {
		ws.Close()
		return
	}
	c := &client{
		supervisor: s,
		ws:         ws,
		runner:     r,
		signal:     make(chan types.Message),
	}
	log.Println("[supervisor] running for: " + namespace)

	c.runner.registerCh <- c
	go c.readLoop()
	go c.writeLoop()
}

func RunSupervisor(ws *websocket.Conn, namespace string) {
	runClient(true, ws, namespace)
}

func RunParticipant(ws *websocket.Conn, namespace string) {
	runClient(false, ws, namespace)
}
