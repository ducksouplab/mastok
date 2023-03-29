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

func (r *runner) tickStateMessage() {
	if r.updateStateTicker != nil {
		r.updateStateTicker.stop()
	}
	ticker := newTicker(models.SessionDurationUnit)
	go ticker.loop(r)
	r.updateStateTicker = ticker
}

func (r *runner) isDone() chan struct{} {
	return r.doneCh
}

func (r *runner) stop() {
	deleteRunner(r.campaign)
	close(r.doneCh)
}

// when process* methods return true, the runner loop is supposed to be stopped
func (r *runner) processRegister(target *client) (done bool) {
	if r.campaign.State == "Running" || target.isSupervisor {
		r.clients.add(target)
		target.outgoingCh <- stateMessage(r.campaign, target)

		if target.isSupervisor {
			// only inform supervisor client about the room size right away
			target.outgoingCh <- poolSizeMessage(r)
		}
	} else {
		// don't register
		target.outgoingCh <- stateMessage(r.campaign, target)
		target.outgoingCh <- disconnectMessage()
		if r.clients.isEmpty() {
			r.stop()
			return true
		}
	}
	return false
}

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

func (r *runner) processLand(target *client, fingerprint string) (done bool) {
	var isInLiveSession bool
	participation, hasParticipated := models.GetParticipation(*r.campaign, fingerprint)
	if hasParticipated {
		pastSession, ok := models.GetSession(participation.SessionID)
		if ok {
			isInLiveSession = pastSession.IsLive()
		}
	}
	if isInLiveSession { // we assume it's a reconnect, so we redirect to oTree
		if done := r.processUnregisterWithReason(target, landRedirectMessage(participation.OtreeCode)); done {
			return true
		}
		return false
	}
	if r.campaign.JoinOnce {
		_, isAlreadyThere := r.roomFingerprints[fingerprint]
		if isAlreadyThere || hasParticipated {
			if done := r.processUnregisterWithReason(target, landRejectMessage()); done {
				return true
			}
			return false
		}
		r.roomFingerprints[fingerprint] = true
	}
	// finally lands in room
	target.outgoingCh <- consentMessage(r.campaign)
	return false
}

func (r *runner) processTentativeJoin(target *client) (done bool) {
	addedToPool, addedToPending := r.clients.tentativeJoin(target)
	if !addedToPool {
		if addedToPending {
			for c := range r.clients.supervisors {
				c.outgoingCh <- pendingSizeMessage(r)
			}
			return false
		} else {
			if done := r.processUnregisterWithReason(target, roomFullMessage()); done {
				return true
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
	return false
}

func (r *runner) processStartSession() (done bool) {
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
				return true
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
	return false
}

func (r *runner) processState(state string) (done bool) {
	r.campaign.State = state
	models.DB.Save(r.campaign)
	for c := range r.clients.all {
		newMessageState := stateMessage(r.campaign, c)
		c.outgoingCh <- newMessageState
		if newMessageState.Payload == "Unavailable" {
			if done := r.processUnregisterWithReason(c, disconnectMessage()); done {
				return true
			}
		}
	}
	return false
}

func (r *runner) loop() {
	for {
		select {
		case c := <-r.registerCh:
			if done := r.processRegister(c); done {
				return
			}
		case c := <-r.unregisterCh:
			if done := r.processUnregister(c); done {
				return
			}
		case m := <-r.participantCh:
			if m.Kind == "Land" {
				if done := r.processLand(m.From, m.Payload); done {
					return
				}
			} else if m.Kind == "Choose" {
				if done := r.processTentativeJoin(m.From); done {
					return
				}
				// starts session when there is valid pool pending
				ready := r.clients.isPoolFull() && !r.campaign.IsBusy()
				if ready {
					if done := r.processStartSession(); done {
						return
					}
				}
			}
		case m := <-r.supervisorCh:
			state := m.Payload.(string)
			if done := r.processState(state); done {
				return
			}
		}
	}
}
