package live

import (
	"log"
	"time"

	"github.com/ducksouplab/mastok/env"
	"github.com/ducksouplab/mastok/models"
)

var checkStartPeriod time.Duration

func init() {
	if env.Mode == "TEST" {
		checkStartPeriod = 3 * time.Millisecond
	} else {
		checkStartPeriod = 1 * time.Second
	}
}

// clients hold references to any (supervisor or participant) client
type runner struct {
	// configuration
	campaign *models.Campaign
	grouping *models.Grouping // ready only, cached from campaign
	// state
	state            string
	clients          *runnerClients
	roomFingerprints map[string]bool
	// manage broadcasting
	registerCh   chan *client
	unregisterCh chan *client
	// other incoming events
	participantCh chan FromParticipantMessage
	supervisorCh  chan Message
	// ticker
	checkStartTicker *time.Ticker
	// signals end
	doneCh chan struct{}
}

// Busy is a temporary state, participants can wait
func isOn(state string) bool {
	return state == models.Running || state == models.Busy
}

func newRunner(c *models.Campaign) *runner {
	group := c.GetGrouping()
	state := c.State

	r := runner{
		campaign:         c,
		grouping:         group,
		state:            state,
		clients:          newRunnerClients(c, group),
		roomFingerprints: map[string]bool{}, // used only for JoinOnce campaigns
		registerCh:       make(chan *client),
		unregisterCh:     make(chan *client),
		// messages coming from participants or supervisor
		participantCh: make(chan FromParticipantMessage),
		supervisorCh:  make(chan Message),
		// ticker
		checkStartTicker: time.NewTicker(checkStartPeriod),
		// signals end
		doneCh: make(chan struct{}),
	}

	if state != models.Running && state != models.Busy {
		r.stopTicker()
	}

	return &r
}

func (r *runner) startTicker() {
	r.checkStartTicker.Stop()
	r.checkStartTicker = time.NewTicker(checkStartPeriod)
}

func (r *runner) stopTicker() {
	r.checkStartTicker.Stop()
}

func (r *runner) isDone() chan struct{} {
	return r.doneCh
}

func (r *runner) stop() {
	r.stopTicker()
	deleteRunner(r.campaign)
	close(r.doneCh)
}

// when process* methods return true, the runner loop is supposed to be stopped
func (r *runner) processRegister(target *client) (done bool) {
	if isOn(r.state) || target.isSupervisor {
		r.clients.add(target)

		if target.isSupervisor {
			target.outgoingCh <- stateMessage(r.campaign.GetLiveState()) // can be busy
			// only inform supervisor client about the room size right away
			target.outgoingCh <- poolSizeMessage(r)
			target.outgoingCh <- pendingSizeMessage(r)
		} else {
			target.outgoingCh <- stateMessage(r.campaign.State)
		}
	} else {
		// don't register
		target.outgoingCh <- stateMessage(models.Unavailable)
		target.outgoingCh <- disconnectMessage("Unavailable")
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
		r.clients.addOneToPoolFromPending()
		// tells the pool and supervisors that the room size has changed
		for c := range r.clients.pool {
			c.outgoingCh <- poolSizeMessage(r)
		}
		for c := range r.clients.supervisors {
			c.outgoingCh <- poolSizeMessage(r)
			c.outgoingCh <- pendingSizeMessage(r)
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
		if done := r.processUnregisterWithReason(target, disconnectMessage("Redirect:"+participation.OtreeCode)); done {
			return true
		}
		return false
	}
	if r.campaign.JoinOnce {
		_, isAlreadyThere := r.roomFingerprints[fingerprint]
		if isAlreadyThere || hasParticipated {
			if done := r.processUnregisterWithReason(target, disconnectMessage("LandingFailed")); done {
				return true
			}
			return false
		}
		r.roomFingerprints[fingerprint] = true
	}
	// finally lands
	target.outgoingCh <- consentMessage(r.campaign)
	return false
}

func (r *runner) processTentativeJoin(target *client) (done bool) {
	addedToPool, addedToPending := r.clients.tentativeJoin(target)
	if !addedToPool {
		if addedToPending {
			target.outgoingCh <- pendingMessage()
			for c := range r.clients.supervisors {
				c.outgoingCh <- pendingSizeMessage(r)
			}
			return false
		} else {
			if done := r.processUnregisterWithReason(target, disconnectMessage("Full")); done {
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

func (r *runner) processState(newState string) (done bool) {
	// persist
	r.campaign.State = newState
	models.DB.Save(r.campaign)
	// notice supervisors
	for c := range r.clients.supervisors {
		c.outgoingCh <- stateMessage(newState)
	}
	// may turn on runner
	liveState := r.campaign.GetLiveState()
	on := isOn(liveState)
	oldOn := isOn(r.state)
	if !oldOn && on {
		r.startTicker()
	}
	// notice or disconnect participants
	for c := range r.clients.participants {
		if on {
			c.outgoingCh <- stateMessage(newState)
		} else {
			c.outgoingCh <- stateMessage(models.Unavailable)
			if done := r.processUnregisterWithReason(c, disconnectMessage("Unavailable")); done {
				return true
			}
		}
	}
	return false
}

func (r *runner) checkIfStillBusy() {
	if !r.campaign.IsBusy() {
		newState := r.campaign.State
		r.state = newState
		for c := range r.clients.supervisors {
			c.outgoingCh <- stateMessage(newState)
		}
	}
}

func (r *runner) processIfPoolReady() (done bool) {
	// check if there is a valid pool pending
	ready := r.clients.isPoolFull() && !r.campaign.IsBusy()
	if !ready {
		// update supervisors clients with state if has changed
		if r.state != r.campaign.GetLiveState() {
			r.state = r.campaign.GetLiveState()
			for c := range r.clients.supervisors {
				c.outgoingCh <- stateMessage(r.campaign.GetLiveState()) // may not be Busy anymore
			}
		}
		return false
	}
	// start session
	newSession, participantCodes, err := models.CreateSession(r.campaign)
	if err != nil {
		log.Println("[runner] session creation failed: ", err)
		return false
	}
	// send SessionStart with oTree URL forged with a unique code
	participantIndex := 0
	var inSession []*client
	for c := range r.clients.pool {
		inSession = append(inSession, c)
		code := participantCodes[participantIndex]
		c.outgoingCh <- sessionStartParticipantMessage(code)
		models.CreateParticipation(newSession, c.fingerprint, code)
		participantIndex++
	}
	// update state (may become Busy or Completed)
	r.state = r.campaign.GetLiveState()
	// notice supervisors
	for c := range r.clients.supervisors {
		c.outgoingCh <- sessionStartSupervisorMessage(newSession)
		c.outgoingCh <- stateMessage(r.state) // if state becomes Busy
	}
	// empty the pool (unregister will fill the pool from pending if possible)
	for _, c := range inSession {
		if done := r.processUnregisterWithReason(c, disconnectMessage("Start")); done {
			return true
		}
	}

	return false
}

func (r *runner) loop() {
	for {
		select {
		case <-r.checkStartTicker.C:
			if r.state == models.Busy {
				r.checkIfStillBusy()
			} else if r.state == models.Running {
				if done := r.processIfPoolReady(); done {
					return
				}
			} else { // Paused or Completed
				r.stopTicker()
			}
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
			} else if m.Kind == "Connect" {
				if done := r.processTentativeJoin(m.From); done {
					return
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
