package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionArgs_Unit(t *testing.T) {
	t.Run("oTree session id follows the mk:namespace:#session format", func(t *testing.T) {
		ns := "fxt_models_ns1"
		campaign, err := FindCampaignByNamespace(ns)
		if err != nil {
			t.Error(err)
		}

		// the fixture data is what we expected
		assert.Equal(t, 3, campaign.StartedSessions)
		assert.Equal(t, "Running", campaign.State)

		// NO, smaller function
		sessionArgs := newSessionArgs(campaign)
		assert.Equal(t, "mk:fxt_models_ns1:4", sessionArgs.Config.Id)
	})
}
