package live

import (
	"log"

	"github.com/ducksouplab/mastok/models"
)

type Message struct {
	Kind    string `json:"kind"`
	Payload any    `json:"payload"`
}

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
	messageCh chan Message
}

func (c *client) write(m Message) {
	if err := c.ws.WriteJSON(m); err != nil {
		log.Println(err)
	}
}

func (c *client) read() (m Message, err error) {
	err = c.ws.ReadJSON(&m)
	return
}

func (c *client) stop() {
	c.runner.unregisterCh <- c
}

func (c *client) readLoop() {
	defer c.stop()

	for {
		m, err := c.read()

		if err != nil {
			log.Println("[ws] ReadJSON error:", err)
			return
		} else if m.Kind == "State" && c.isSupervisor {
			c.runner.stateCh <- m.Payload.(string)
		}
	}
}

// at most one writer to a connection since all writes happen in this goroutine
// like in https://github.com/gorilla/websocket/blob/master/examples/chat/client.go
func (c *client) writeLoop() {
	defer c.stop()
	for m := range c.messageCh {
		c.write(m)
		if m.Kind == "Participant" && m.Payload == "Disconnect" {
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
		messageCh:    make(chan Message, 256),
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
