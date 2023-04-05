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
		Payload: strconv.Itoa(r.clients.pendingSize()) + "/" + strconv.Itoa(r.clients.maxPending),
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

func pendingMessage() Message {
	return Message{
		Kind: "Pending",
	}
}

func consentMessage(c *models.Campaign) Message {
	return Message{
		Kind:    "Consent",
		Payload: helpers.MarkdownToHTML(c.Consent),
	}
}

func groupingMessage(c *models.Campaign) Message {
	return Message{
		Kind:    "Grouping",
		Payload: c.Grouping,
	}
}

// Disconnect messages below

// campaign is not running or after SessionStart
func disconnectMessage(reason string) Message {
	return Message{
		Kind:    "Disconnect",
		Payload: reason,
	}
}
