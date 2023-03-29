package live

import (
	"log"

	"github.com/ducksouplab/mastok/helpers"
	"github.com/ducksouplab/mastok/models"
)

type Message struct {
	Kind    string `json:"kind"`
	Payload any    `json:"payload"`
}

type FromParticipantMessage struct {
	Kind    string
	Payload string
	From    *client
}

type wsConn interface {
	ReadJSON(any) error
	WriteJSON(any) error
	Close() error
}

// client for campaign runner
type client struct {
	// state
	isSupervisor bool
	hasLanded    bool
	hasAgreed    bool
	fingerprint  string
	groupLabel   string
	// links
	ws     wsConn
	runner *runner
	// updates from runner
	outgoingCh chan Message
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

func (c *client) unregister() {
	c.runner.unregisterCh <- c
}

func (c *client) readLoop() {
	defer c.unregister()

	for {
		m, err := c.read()

		if err != nil {
			// client left (or needs to stop loop anyway)
			return
		} else if c.isSupervisor {
			if m.Kind == "State" {
				c.runner.supervisorCh <- m
			}
		} else { // participant
			if m.Kind == "Land" {
				fingerprint := m.Payload.(string)
				if len(fingerprint) == 0 {
					c.outgoingCh <- landRejectMessage()
					break
				} else {
					// do set client state before sharing landing with runner
					c.hasLanded = true
					c.fingerprint = fingerprint
					c.runner.participantCh <- FromParticipantMessage{Kind: m.Kind, Payload: fingerprint, From: c}
				}
			} else if m.Kind == "Agree" {
				if c.hasLanded { // can't agree before landing
					c.hasAgreed = true
					c.runner.participantCh <- FromParticipantMessage{Kind: m.Kind, From: c}
					// when there is not grouping, Agree implies Choose
					if c.runner.grouping == nil {
						c.groupLabel = defaultGroupLabel
						c.runner.participantCh <- FromParticipantMessage{Kind: "Choose", Payload: defaultGroupLabel, From: c}
					}
				}
			} else if m.Kind == "Choose" {
				groupLabel := m.Payload.(string)
				if len(groupLabel) != 0 && c.hasAgreed { // can't agree before landing
					c.groupLabel = groupLabel
					c.runner.participantCh <- FromParticipantMessage{Kind: m.Kind, Payload: groupLabel, From: c}
				}
			}
		}
	}
}

// at most one writer to a connection since all writes happen in this goroutine
// like in https://github.com/gorilla/websocket/blob/master/examples/chat/client.go
func (c *client) writeLoop() {
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
	var ok bool
	if isSupervisor {
		campaign, ok = models.GetCampaignByNamespace(identifier)
	} else {
		campaign, ok = models.GetCampaignBySlug(identifier)
	}
	if !ok {
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
