package live

import (
	"log"

	"github.com/ducksouplab/mastok/helpers"
	"github.com/ducksouplab/mastok/models"
	"github.com/ducksouplab/mastok/otree"
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
	// state
	isSupervisor  bool
	hasLanded     bool
	hasJoinedPool bool
	fingerprint   string
	// links
	ws     wsConn
	runner *runner
	// updates from runner
	outgoingCh chan Message
}

func redirectMessage(code string) Message {
	return Message{
		Kind:    "Redirect",
		Payload: otree.ParticipantStartURL(code),
	}
}

func rejectMessage() Message {
	return Message{
		Kind: "Reject",
	}
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
		} else if c.isSupervisor {
			if m.Kind == "State" {
				c.runner.incomingCh <- m
			}
		} else { // participant
			if m.Kind == "Land" {
				fingerprint := m.Payload.(string)

				if len(fingerprint) != 0 {
					// set client state
					c.fingerprint = fingerprint
					c.hasLanded = true
					// process if reply is needed
					// p, err := models.FindParticipation(*c.runner.campaign, fingerprint)
					// log.Printf(">> >> p %#v %#v", p, err)
					// if err == nil {
					// 	s, err := models.FindSession(p.SessionID)
					// 	log.Printf(">> >> s %#v %#v %#v", s, s.IsLive(), err)
					// 	if err == nil {
					// 		if s.IsLive() {
					// 			c.outgoingCh <- redirectMessage(p.OtreeCode)
					// 		} else {
					// 			c.outgoingCh <- rejectMessage()
					// 		}
					// 	}
					// }
				}
			} else if m.Kind == "Join" {
				if c.hasLanded {
					c.runner.incomingCh <- m
					c.hasJoinedPool = true
				}
			}
		}
	}
}

// at most one writer to a connection since all writes happen in this goroutine
// like in https://github.com/gorilla/websocket/blob/master/examples/chat/client.go
func (c *client) writeLoop() {
	defer c.stop()
	for m := range c.outgoingCh {
		c.write(m)
		if helpers.Contains([]string{"Disconnect", "Redirect", "Reject"}, m.Kind) {
			// stops for loop
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
		outgoingCh:   make(chan Message, 256),
	}
	log.Println("[client] running for: " + identifier)

	go c.readLoop()
	go c.writeLoop()

	// participants have to Land first
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
