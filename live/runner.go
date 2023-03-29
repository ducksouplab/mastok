package live

import (
	"log"

	"github.com/ducksouplab/mastok/models"
)

const maxPendingSize = 50

// clients hold references to any (supervisor or participant) client
type runner struct {
	campaign         *models.Campaign
	grouping         *models.Grouping // ready only, cached from campaign
	clients          *runnerClients
	roomFingerprints map[string]bool
	// manage broadcasting
	registerCh   chan *client
	unregisterCh chan *client
	// other incoming events
	participantCh chan FromParticipantMessage
	supervisorCh  chan Message
	// done
	updateStateTicker *ticker
	doneCh            chan struct{}
}

func newRunner(c *models.Campaign) *runner {
	g := c.GetGrouping()

	r := runner{
		campaign:         c,
		grouping:         g,
		clients:          newRunnerClients(c, g),
		roomFingerprints: map[string]bool{}, // used only for JoinOnce campaigns
		registerCh:       make(chan *client),
		unregisterCh:     make(chan *client),
		// messages coming from participants or supervisor
		participantCh: make(chan FromParticipantMessage),
		supervisorCh:  make(chan Message),
		// signals end
		doneCh: make(chan struct{}),
	}

	if c.GetPublicState(true) == models.Busy {
		r.tickStateMessage()
	}

	return &r
}

func (r *runner) isDone() chan struct{} {
	return r.doneCh
}

func (r *runner) stop() {
	deleteRunner(r.campaign)
	close(r.doneCh)
}

// return value: true to stop runner loop
func (r *runner) processUnregister(target *client) (done bool) {
	// deletes client
	if wasInPool := r.clients.delete(target); wasInPool {
		// tells everyone including supervisor that the room size has changed
		for c := range r.clients.pool {
			c.outgoingCh <- poolSizeMessage(r)
		}
		for c := range r.clients.supervisors {
			c.outgoingCh <- poolSizeMessage(r)
		}
	} else {
		for c := range r.clients.supervisors {
			c.outgoingCh <- pendingSizeMessage(r)
		}
	}
	if r.campaign.JoinOnce {
		delete(r.roomFingerprints, target.fingerprint)
	}
	if r.clients.isEmpty() {
		r.stop()
		return true
	}
	return false
}

func (r *runner) processUnregisterWithReason(target *client, m Message) (done bool) {
	target.outgoingCh <- m
	return r.processUnregister(target)
}

func (r *runner) loop() {
	for {
		select {
		case c := <-r.registerCh:
			if r.campaign.State == "Running" || c.isSupervisor {
				r.clients.add(c)
				c.outgoingCh <- stateMessage(r.campaign, c)

				if c.isSupervisor {
					// only inform supervisor client about the room size right away
					c.outgoingCh <- poolSizeMessage(r)
				}
			} else {
				// don't register
				c.outgoingCh <- stateMessage(r.campaign, c)
				c.outgoingCh <- disconnectMessage()
				if r.clients.isEmpty() {
					r.stop()
					return
				}
			}
		case c := <-r.unregisterCh:
			if done := r.processUnregister(c); done {
				return
			}
		case m := <-r.participantCh:
			if m.Kind == "Land" {
				var isInLiveSession bool
				participation, hasParticipated := models.GetParticipation(*r.campaign, m.Payload)
				if hasParticipated {
					pastSession, ok := models.GetSession(participation.SessionID)
					if ok {
						isInLiveSession = pastSession.IsLive()
					}
				}
				if isInLiveSession { // we assume it's a reconnect, so we redirect to oTree
					if done := r.processUnregisterWithReason(m.From, landRedirectMessage(participation.OtreeCode)); done {
						return
					}
					break
				}
				if r.campaign.JoinOnce {
					_, isAlreadyThere := r.roomFingerprints[m.Payload]
					if isAlreadyThere || hasParticipated {
						if done := r.processUnregisterWithReason(m.From, landRejectMessage()); done {
							return
						}
						break
					}
					r.roomFingerprints[m.Payload] = true
				}
				// finally lands in room
				m.From.outgoingCh <- consentMessage(r.campaign)
			} else if m.Kind == "Choose" {
				addedToPool, addedToPending := r.clients.tentativeJoin(m.From)
				if !addedToPool {
					if addedToPending {
						for c := range r.clients.supervisors {
							c.outgoingCh <- pendingSizeMessage(r)
						}
						break
					} else {
						if done := r.processUnregisterWithReason(m.From, roomFullMessage()); done {
							return
						}
					}
				}
				// inform everyone (participants in pool and supervisors) about the new room size
				for c := range r.clients.pool {
					c.outgoingCh <- poolSizeMessage(r)
				}
				for c := range r.clients.supervisors {
					c.outgoingCh <- poolSizeMessage(r)
				}
				// starts session when there is valid pool pending
				ready := r.clients.isPoolFull()
				if ready {
					session, participantCodes, err := models.CreateSession(r.campaign)
					if err != nil {
						log.Println("[runner] session creation failed: ", err)
					} else {
						participantIndex := 0
						for c := range r.clients.pool {
							code := participantCodes[participantIndex]
							c.outgoingCh <- sessionStartParticipantMessage(code)
							models.CreateParticipation(session, c.fingerprint, code)
							if done := r.processUnregisterWithReason(c, disconnectMessage()); done {
								return
							}
							participantIndex++
						}

						for c := range r.clients.supervisors {
							c.outgoingCh <- sessionStartSupervisorMessage(session)
							c.outgoingCh <- stateMessage(r.campaign, c) // if state becomes Busy
							if c.runner.campaign.GetPublicState(true) == models.Busy {
								c.runner.tickStateMessage()
							}
						}
						if updated := r.clients.resetPoolFromPending(); updated {
							// now pool has been emptied, refill it from pending participants
							for c := range r.clients.pool { // it's a new pool
								c.outgoingCh <- poolSizeMessage(r)
							}
						}
						for c := range r.clients.supervisors {
							c.outgoingCh <- poolSizeMessage(r)
							c.outgoingCh <- pendingSizeMessage(r)
						}
					}
				}
			}
		case m := <-r.supervisorCh:
			if m.Kind == "State" {
				r.campaign.State = m.Payload.(string)
				models.DB.Save(r.campaign)
				for c := range r.clients.all {
					newMessageState := stateMessage(r.campaign, c)
					c.outgoingCh <- newMessageState
					if newMessageState.Payload == "Unavailable" {
						if done := r.processUnregisterWithReason(c, disconnectMessage()); done {
							return
						}
					}
				}
			}
		}
	}
}
