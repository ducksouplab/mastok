package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCampaign_Unit(t *testing.T) {
	t.Run("session can't be added to completed campaign", func(t *testing.T) {
		ns := "fxt_models_ns2_completed"
		campaign, _ := FindCampaignByNamespace(ns)

		session := Session{}
		err := campaign.appendSession(session)

		assert.Error(t, err)
	})

	t.Run("campaign is Busy when MaxSessions is reached", func(t *testing.T) {
		ns := "fxt_models_ns3_busy"
		campaign, _ := FindCampaignByNamespace(ns)

		assert.Equal(t, 3, campaign.liveSessions())
		assert.Equal(t, true, campaign.isBusy())
		assert.Equal(t, Busy, campaign.GetPublicState(true))
		assert.Equal(t, Unavailable, campaign.GetPublicState(false))
	})
}
