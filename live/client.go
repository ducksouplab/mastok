package live

import (
	"log"

	"github.com/gorilla/websocket"
)

// client for campaign runner
type client struct {
	supervisor bool
	ws         *wsConn
	runner     *runner
	signal     chan string
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
		} else if m.Kind == "state" {
			c.runner.stateCh <- m.Payload
		} else if m.Kind == "join" {
			c.runner.joinCh <- m.Payload
		}
	}
}

func (c *client) writeLoop() {
	//s.ws.write(message{Kind: "kind"})
	for {
		select {
		case state := <-c.runner.stateCh:
			log.Println(state)
		case join := <-c.runner.joinCh:
			log.Println(join)
		case newSession := <-c.runner.newSessionCh:
			log.Println(newSession)
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
		signal:     make(chan string),
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
