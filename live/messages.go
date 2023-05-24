package live

import (
	"strconv"

	"github.com/ducksouplab/mastok/helpers"
	"github.com/ducksouplab/mastok/models"
	"github.com/ducksouplab/mastok/otree"
)

// in case there is no grouping
const defaultGroupLabel = "default"

// Participant or supervisor messages

func stateMessage(state string) Message {
	return Message{
		Kind:    "State",
		Payload: state,
	}
}

func joiningSizeMessage(r *runner) Message {
	return Message{
		Kind:    "JoiningSize",
		Payload: strconv.Itoa(r.clients.joiningSize()) + "/" + strconv.Itoa(r.campaign.PerSession),
	}
}

// Supervisor messages

func pendingSizeMessage(r *runner) Message {
	return Message{
		Kind:    "PendingSize",
		Payload: strconv.Itoa(r.clients.pendingSize()) + "/" + strconv.Itoa(r.clients.maxPending),
	}
}

func sessionStartMessage(session models.Session) Message {
	return Message{
		Kind:    "SessionStart",
		Payload: session,
	}
}

// Participant messages

func startingMessage(code string) Message {
	return Message{
		Kind:    "Starting",
		Payload: otree.ParticipantStartURL(code),
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

func instructionsMessage(c *models.Campaign) Message {
	return Message{
		Kind:    "Instructions",
		Payload: helpers.MarkdownToHTML(c.Instructions),
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
