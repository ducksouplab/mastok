package live

import (
	"log"
	"strings"

	"github.com/ducksouplab/mastok/models"
)

type wsConn interface {
	ReadJSON(any) error
	WriteJSON(any) error
	Close() error
}

// client for campaign runner
type client struct {
	isSupervisor bool
	ws           wsConn
	runner       *runner
	// updates from runner
	signalCh chan string
}

func (c *client) write(signal string) {
	if err := c.ws.WriteJSON(signal); err != nil {
		log.Println(err)
	}
}

func (c *client) read() (kind, payload string, err error) {
	var m string
	err = c.ws.ReadJSON(&m)
	if err != nil {
		return
	}
	slice := strings.Split(m, ":")
	kind = slice[0]
	payload = slice[1]
	return
}

func (c *client) stop() {
	c.runner.unregisterCh <- c
}

func (c *client) readLoop() {
	defer c.stop()

	for {
		kind, payload, err := c.read()

		if err != nil {
			log.Println("[ws] ReadJSON error:", err)
			return
		} else if kind == "State" && c.isSupervisor {
			c.runner.stateCh <- payload
		}
	}
}

// at most one writer to a connection since all writes happen in this goroutine
// like in https://github.com/gorilla/websocket/blob/master/examples/chat/client.go
func (c *client) writeLoop() {
	defer c.stop()
	for signal := range c.signalCh {
		c.write(signal)
		if signal == "Participant:Disconnect" {
			return
		}
	}
}

func runClient(isSupervisor bool, ws wsConn, identifier string) *client {
	var campaign *models.Campaign
	var err error
	if isSupervisor {
		campaign, err = models.FindCampaignByNamespace(identifier)
	} else {
		campaign, err = models.FindCampaignBySlug(identifier)
	}
	if err != nil {
		return nil
	}

	r, err := getRunner(campaign)
	if err != nil {
		ws.Close()
		return nil
	}
	c := &client{
		isSupervisor: isSupervisor,
		ws:           ws,
		runner:       r,
		signalCh:     make(chan string, 256),
	}
	log.Println("[client] running for: " + identifier)

	go c.readLoop()
	go c.writeLoop()

	c.runner.registerCh <- c
	return c
}

// identified by namespace
func RunSupervisor(ws wsConn, namespace string) *client {
	return runClient(true, ws, namespace)
}

// identified by slug
func RunParticipant(ws wsConn, slug string) *client {
	return runClient(false, ws, slug)
}
