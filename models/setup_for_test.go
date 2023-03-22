package models

import (
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ducksouplab/mastok/env"
)

var sessionDurationTest = 60

var Fixtures []Campaign = []Campaign{
	{
		Namespace:          "fxt_models_ns1",
		Slug:               "fxt_models_ns1_slug",
		OtreeExperiment:    "xp_name",
		PerSession:         4,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		StartedSessions:    3,
		State:              Running,
	},
	{
		Namespace:          "fxt_models_ns2_completed",
		Slug:               "fxt_models_ns2_completed_slug",
		OtreeExperiment:    "xp_name",
		PerSession:         4,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		StartedSessions:    4,
		State:              Completed,
	},
	{
		Namespace:          "fxt_models_ns3_busy",
		Slug:               "fxt_models_ns3_busy_slug",
		OtreeExperiment:    "xp_name",
		PerSession:         4,
		MaxSessions:        8,
		ConcurrentSessions: 3,
		SessionDuration:    sessionDurationTest,
		StartedSessions:    3,
		State:              Running,
	},
}

func TestMain(m *testing.M) {
	ReinitTestDB()
	if err := DB.Create(Fixtures).Error; err != nil {
		log.Fatal(err)
	}
	busyCampaign, _ := FindCampaignByNamespace("fxt_models_ns3_busy")
	var oldSession = Session{
		LaunchedAt: time.Now().Add(-2 * time.Duration(sessionDurationTest) * SessionDurationUnit), // not "current" session
		Code:       "codenum1",
		OtreeId:    "mk:fxt_models_ns3_busy:1",
		Size:       4,
		AdminUrl:   env.OTreeURL + "/SessionStartLinks/codenum1",
	}
	busyCampaign.appendSession(&oldSession)
	for i := 0; i < 3; i++ {
		// less than sessionDurationTest
		ago := -time.Duration(sessionDurationTest*(i+1)/4) * SessionDurationUnit
		suffix := strconv.Itoa(i + 2)
		currentSession := Session{
			LaunchedAt: time.Now().Add(ago),
			Code:       "codenum" + suffix,
			OtreeId:    "mk:fxt_models_ns3_busy:" + suffix,
			Size:       4,
			AdminUrl:   env.OTreeURL + "/SessionStartLinks/codenum" + suffix,
		}
		busyCampaign.appendSession(&currentSession)
	}

	os.Exit(m.Run())
}
