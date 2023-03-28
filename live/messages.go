package live

import (
	"strconv"

	"github.com/ducksouplab/mastok/helpers"
	"github.com/ducksouplab/mastok/models"
	"github.com/ducksouplab/mastok/otree"
)

// in case there is no grouping
const defaultGroupLabel = "default"

func stateMessage(c *models.Campaign, cl *client) Message {
	return Message{
		Kind:    "State",
		Payload: c.GetPublicState(cl.isSupervisor),
	}
}

func poolSizeMessage(r *runner) Message {
	return Message{
		Kind:    "PoolSize",
		Payload: strconv.Itoa(r.clients.participantsCount()) + "/" + strconv.Itoa(r.campaign.PerSession),
	}
}

func sessionStartParticipantMessage(code string) Message {
	return Message{
		Kind:    "SessionStart",
		Payload: otree.ParticipantStartURL(code),
	}
}

func sessionStartSupervisorMessage(session models.Session) Message {
	return Message{
		Kind:    "SessionStart",
		Payload: session,
	}
}

func participantDisconnectMessage() Message {
	return Message{
		Kind: "Disconnect",
	}
}

func participantRejectMessage() Message {
	return Message{
		Kind: "Reject",
	}
}

func participantRedirectMessage(code string) Message {
	return Message{
		Kind:    "Redirect",
		Payload: otree.ParticipantStartURL(code),
	}
}

func participantConsentMessage(c *models.Campaign) Message {
	return Message{
		Kind:    "Consent",
		Payload: helpers.MarkdownToHTML(c.Consent),
	}
}

func (r *runner) tickStateMessage() {
	if r.updateStateTicker != nil {
		r.updateStateTicker.stop()
	}
	ticker := newTicker(models.SessionDurationUnit)
	go ticker.loop(r)
	r.updateStateTicker = ticker
}
