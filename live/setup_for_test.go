package live

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/ducksouplab/mastok/models"
)

const (
	shortDuration       = 10 * time.Millisecond
	longDuration        = 50 * time.Millisecond  // for instance if there are DB writes
	longerDuration      = 120 * time.Millisecond // for instance if there are DB writes
	sessionDurationTest = 60
)

func TestMain(m *testing.M) {
	models.ReinitTestDB()
	if err := models.DB.Create(Fixtures).Error; err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
}

var Fixtures []models.Campaign = []models.Campaign{
	{
		Namespace:          "fxt_live_ns1",
		Slug:               "fxt_live_ns1_slug",
		OtreeExperiment:    "xp_name",
		PerSession:         4,
		MaxSessions:        2,
		ConcurrentSessions: 2,
		State:              models.Running,
	},
	{
		Namespace:          "fxt_live_ns2_to_be_paused",
		Slug:               "fxt_live_ns2_to_be_paused_slug",
		OtreeExperiment:    "xp_name",
		PerSession:         4,
		MaxSessions:        2,
		ConcurrentSessions: 2,
		State:              models.Running,
	},
	{
		Namespace:          "fxt_live_ns3_paused",
		Slug:               "fxt_live_ns3_paused_slug",
		OtreeExperiment:    "xp_name",
		PerSession:         8,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		State:              models.Paused,
	},
	{
		Namespace:          "fxt_live_ns4_paused",
		Slug:               "fxt_live_ns4_paused_slug",
		OtreeExperiment:    "xp_name",
		PerSession:         8,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		State:              models.Paused,
	},
	{
		Namespace:          "fxt_live_ns5_launched",
		Slug:               "fxt_live_ns5_launched_slug",
		OtreeExperiment:    "xp_name",
		PerSession:         4,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		State:              models.Running,
	},
	{
		Namespace:          "fxt_live_ns6_completed",
		Slug:               "fxt_live_ns6_completed_slug",
		OtreeExperiment:    "xp_name",
		PerSession:         8,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		State:              models.Completed,
	},
	{
		Namespace:          "fxt_live_ns7_almost_completed",
		Slug:               "fxt_live_ns7_almost_completed_slug",
		OtreeExperiment:    "xp_name",
		PerSession:         4,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		StartedSessions:    3,
		State:              models.Running,
	},
	{
		Namespace:          "fxt_live_ns8_busy",
		Slug:               "fxt_live_ns8_busy_slug",
		OtreeExperiment:    "xp_name",
		PerSession:         2,
		MaxSessions:        2,
		ConcurrentSessions: 1,
		SessionDuration:    sessionDurationTest,
		State:              models.Running,
	},
	{
		Namespace:          "fxt_live_ns9_once",
		Slug:               "fxt_live_ns9_once_slug",
		OtreeExperiment:    "xp_name",
		PerSession:         4,
		JoinOnce:           true,
		MaxSessions:        2,
		ConcurrentSessions: 2,
		State:              models.Running,
	},
	{
		Namespace:          "fxt_live_ns10_redirect",
		Slug:               "fxt_live_ns10_redirect_slug",
		OtreeExperiment:    "xp_name",
		PerSession:         2,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		StartedSessions:    0,
		SessionDuration:    sessionDurationTest,
		State:              models.Running,
	},
	{
		Namespace:          "fxt_live_ns11_redirect2",
		Slug:               "fxt_live_ns11_redirect2_slug",
		OtreeExperiment:    "xp_name",
		PerSession:         2,
		JoinOnce:           true,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		StartedSessions:    0,
		SessionDuration:    sessionDurationTest,
		State:              models.Running,
	},
}
