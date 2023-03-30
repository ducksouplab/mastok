package live

import (
	"strconv"

	"github.com/ducksouplab/mastok/helpers"
	"github.com/ducksouplab/mastok/models"
	"github.com/ducksouplab/mastok/otree"
)

// in case there is no grouping
const defaultGroupLabel = "default"

func stateMessage(state string) Message {
	return Message{
		Kind:    "State",
		Payload: state,
	}
}

func poolSizeMessage(r *runner) Message {
	return Message{
		Kind:    "PoolSize",
		Payload: strconv.Itoa(r.clients.poolSize()) + "/" + strconv.Itoa(r.campaign.PerSession),
	}
}

func pendingSizeMessage(r *runner) Message {
	return Message{
		Kind:    "PendingSize",
		Payload: strconv.Itoa(r.clients.pendingSize()) + "/" + strconv.Itoa(maxPendingSize),
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

// campaign is not running or after SessionStart
func disconnectMessage() Message {
	return Message{
		Kind: "Disconnect",
	}
}

func landRejectMessage() Message {
	return Message{
		Kind: "Reject",
	}
}

func landRedirectMessage(code string) Message {
	return Message{
		Kind:    "Redirect",
		Payload: otree.ParticipantStartURL(code),
	}
}

func roomFullMessage() Message {
	return Message{
		Kind: "Full",
	}
}

func consentMessage(c *models.Campaign) Message {
	return Message{
		Kind:    "Consent",
		Payload: helpers.MarkdownToHTML(c.Consent),
	}
}
