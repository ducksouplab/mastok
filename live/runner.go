package live

import (
	"errors"
	"time"

	"github.com/ducksouplab/mastok/env"
	"github.com/ducksouplab/mastok/models"
	"github.com/ducksouplab/mastok/otree"
)

var updatePeriod time.Duration

func init() {
	if env.Mode == "TEST" {
		updatePeriod = 3 * time.Millisecond
	} else {
		updatePeriod = 1 * time.Second
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
	// internal update of campaign
	updateCampaignCh chan *models.Campaign
	// manage broadcasting
	registerCh   chan *client
	unregisterCh chan *client
	// other incoming events
	participantCh chan FromParticipantMessage
	supervisorCh  chan Message
	// ticker
	updateTicker *time.Ticker
	// signals end
	done   bool
	doneCh chan struct{}
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
		updateCampaignCh: make(chan *models.Campaign),
		registerCh:       make(chan *client),
		unregisterCh:     make(chan *client),
		// messages coming from participants or supervisor
		participantCh: make(chan FromParticipantMessage),
		supervisorCh:  make(chan Message),
		// ticker
		updateTicker: time.NewTicker(updatePeriod),
		// signals end
		doneCh: make(chan struct{}),
	}

	if !r.isRunningOrBusy() {
		r.stopTicker()
	}

	return &r
}

// Busy is a temporary state, participants can wait
func (r *runner) isRunningOrBusy() bool {
	return r.state == models.Running || r.state == models.Busy
}

func (r *runner) isAcceptingJoins() bool {
	return r.state != models.Busy && !r.clients.isJoiningFull()
}

func (r *runner) startTicker() {
	r.updateTicker.Stop()
	r.updateTicker = time.NewTicker(updatePeriod)
}

func (r *runner) stopTicker() {
	r.updateTicker.Stop()
}

func (r *runner) isDone() chan struct{} {
	return r.doneCh
}

func (r *runner) stop() {
	if !r.done {
		r.done = true
		r.stopTicker()
		deleteRunner(r.campaign)
		close(r.doneCh)
	}
}

// when process* methods return true, the runner loop is supposed to be stopped
func (r *runner) processRegister(target *client) (done bool) {
	if r.isRunningOrBusy() || target.isSupervisor {
		r.clients.register(target)

		if target.isSupervisor {
			target.outgoingCh <- stateMessage(r.campaign.GetLiveState()) // can be busy
			// only inform supervisor client about the room size right away
			target.outgoingCh <- joiningSizeMessage(r)
			target.outgoingCh <- pendingSizeMessage(r)
		} else {
			target.outgoingCh <- stateMessage(r.campaign.State)
		}
	} else { // don't register
		if r.state == models.Paused {
			target.outgoingCh <- pausedMessage(r.campaign)
		} else if r.state == models.Completed {
			target.outgoingCh <- completedMessage(r.campaign)
		}
		target.outgoingCh <- disconnectMessage(r.state) // should be either Paused, Completed or Unavailable
		if r.clients.isEmpty() {
			r.stop()
			return true
		}
	}
	return false
}

func (r *runner) cleanupLeave(target *client) (done bool) {
	if r.campaign.JoinOnce {
		// doesn't touch participations, only live runner state
		delete(r.roomFingerprints, target.fingerprint)
	}
	if r.clients.isEmpty() {
		r.stop()
		return true
	}
	return false
}

// on js client connection closed
func (r *runner) processUnregisterAndReplace(target *client) (done bool) {
	wasInJoining, wasInPending := r.clients.delete(target)
	if wasInJoining {
		if r.state != models.Busy {
			r.clients.oneTentativeJoinFromPending()
		}
		for c := range r.clients.joining {
			c.outgoingCh <- joiningSizeMessage(r)
		}
		for c := range r.clients.supervisors {
			c.outgoingCh <- joiningSizeMessage(r)
			c.outgoingCh <- pendingSizeMessage(r)
		}
	} else if wasInPending {
		for c := range r.clients.supervisors {
			c.outgoingCh <- pendingSizeMessage(r)
		}
	}
	return r.cleanupLeave(target)
}

// unregister because of start (then updating pool/pending is managed and not instant) or early disconnect
func (r *runner) processManagedUnregister(target *client, m Message) (done bool) {
	target.outgoingCh <- m

	if wasInJoining, wasInPending := r.clients.delete(target); wasInJoining {
		for c := range r.clients.supervisors {
			c.outgoingCh <- joiningSizeMessage(r)
		}
	} else if wasInPending {
		for c := range r.clients.supervisors {
			c.outgoingCh <- pendingSizeMessage(r)
		}
	}
	return r.cleanupLeave(target)
}

func (r *runner) processLand(target *client, fingerprint string) (done bool) {
	participation, hasParticipated := models.GetLastParticipation(*r.campaign, fingerprint)
	// check in case has participated to a session which is currently live
	if env.LiveRedirect && hasParticipated {
		if pastSession, ok := models.GetSession(participation.SessionID); ok {
			if pastSession.IsLive() {
				// we assume it's a reconnect, so we redirect to oTree
				return r.processManagedUnregister(target, disconnectMessage("Redirect:"+otree.ParticipantStartURL(participation.OtreeCode)))
			}
		}
	}
	if r.campaign.JoinOnce {
		_, isAlreadyThere := r.roomFingerprints[fingerprint]
		if isAlreadyThere || hasParticipated {
			return r.processManagedUnregister(target, disconnectMessage("LandingFailed"))
		}
		r.roomFingerprints[fingerprint] = true
	}
	// finally lands
	target.outgoingCh <- consentMessage(r.campaign)
	return false
}

func (r *runner) processTentativeJoinOrPending(target *client) (done bool) {
	// first: try adding to joining
	if r.isAcceptingJoins() && r.clients.tentativeJoin(target) {
		// inform participants in the joining pool and supervisors about the new pool size
		for c := range r.clients.joining {
			c.outgoingCh <- joiningSizeMessage(r)
		}
		for c := range r.clients.supervisors {
			c.outgoingCh <- joiningSizeMessage(r)
		}
		// inform target of instructions
		target.outgoingCh <- instructionsMessage(r.campaign)
		return false
	}
	// second: try adding to pending
	if r.clients.tentativePending(target) {
		target.outgoingCh <- pendingMessage()
		for c := range r.clients.supervisors {
			c.outgoingCh <- pendingSizeMessage(r)
		}
		return false
	}
	// could not be added, unregisters
	return r.processManagedUnregister(target, disconnectMessage("Full"))
}

func (r *runner) processStateUpdate(newState string) (done bool) {
	oldRunningOrBusy := r.isRunningOrBusy()
	r.state = newState
	// persist
	r.campaign.State = newState
	models.DB.Save(r.campaign)
	// notice supervisors
	for c := range r.clients.supervisors {
		c.outgoingCh <- stateMessage(newState)
	}
	// may turn on runner
	liveState := r.campaign.GetLiveState()
	newRunningOrBusy := liveState == models.Running || liveState == models.Busy
	if !oldRunningOrBusy && newRunningOrBusy {
		r.startTicker()
	}
	// notice or disconnect participants
	for c := range r.clients.participants {
		if newRunningOrBusy {
			c.outgoingCh <- stateMessage(newState)
		} else {
			c.outgoingCh <- stateMessage(models.Unavailable)
			if done := r.processManagedUnregister(c, disconnectMessage("Unavailable")); done {
				return true
			}
		}
	}
	return false
}

func (r *runner) updateStateIfNoMoreBusy() {
	if !r.campaign.IsBusy() {
		newState := r.campaign.State
		r.state = newState
		for c := range r.clients.supervisors {
			c.outgoingCh <- stateMessage(newState)
		}
	}
}

func (r *runner) processIfJoiningReady() (done bool, err error) {
	// check if there is a valid pool ready to join
	ready := r.clients.isJoiningFull() && !r.campaign.IsBusy()
	if !ready {
		// update supervisors clients with state if has changed
		if r.state != r.campaign.GetLiveState() {
			r.state = r.campaign.GetLiveState()
			for c := range r.clients.supervisors {
				c.outgoingCh <- stateMessage(r.campaign.GetLiveState()) // may not be Busy anymore
			}
		}
		return false, nil
	}
	// start session
	newSession, participantCodes, err := models.CreateSession(r.campaign)
	if err != nil {
		return false, errors.New("oTree CreateSession failed")
	}
	// send Starting with oTree URL forged with a unique code
	participantIndex := 0
	var inSession []*client
	for c := range r.clients.joining {
		inSession = append(inSession, c)
		code := participantCodes[participantIndex]
		c.outgoingCh <- startingMessage(code)
		models.CreateParticipation(newSession, c.fingerprint, code)
		participantIndex++
	}
	// update state (may become Busy or Completed)
	r.state = r.campaign.GetLiveState()
	// notice supervisors
	for c := range r.clients.supervisors {
		c.outgoingCh <- sessionStartMessage(newSession)
		c.outgoingCh <- stateMessage(r.state) // if state becomes Busy
	}
	// empty the joining pool
	for _, c := range inSession {
		if done := r.processManagedUnregister(c, disconnectMessage("Start")); done {
			return true, nil
		}
	}
	return false, nil
}

func (r *runner) loop() {
	for {
		select {
		case <-r.updateTicker.C:
			if r.state == models.Busy {
				r.updateStateIfNoMoreBusy()
			} else if r.state == models.Running {
				if updated := r.clients.allTentativeJoinFromPending(); updated {
					for c := range r.clients.joining {
						c.outgoingCh <- joiningSizeMessage(r)
					}
					for c := range r.clients.supervisors {
						c.outgoingCh <- joiningSizeMessage(r)
						c.outgoingCh <- pendingSizeMessage(r)
					}
				}
				if done, err := r.processIfJoiningReady(); err != nil {
					r.processStateUpdate(models.Unavailable)
				} else if done {
					return
				}
			} else { // Paused or Completed
				r.stopTicker()
			}
		case campaign := <-r.updateCampaignCh:
			r.campaign = campaign
		case c := <-r.registerCh:
			if done := r.processRegister(c); done {
				return
			}
		case c := <-r.unregisterCh:
			if done := r.processUnregisterAndReplace(c); done {
				return
			}
		case m := <-r.participantCh:
			// m.From is client
			if m.Kind == "Land" {
				if done := r.processLand(m.From, m.Payload); done {
					return
				}
			} else if m.Kind == "Connect" {
				if done := r.processTentativeJoinOrPending(m.From); done {
					return
				}
			}
		case m := <-r.supervisorCh:
			state := m.Payload.(string)
			if done := r.processStateUpdate(state); done {
				return
			}
		}
	}
}
