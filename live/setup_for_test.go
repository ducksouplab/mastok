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
	// supervisor tests
	{
		OtreeExperiment:    "xp_name",
		Namespace:          "fxt_live_sup",
		Slug:               "fxt_live_sup_slug",
		PerSession:         4,
		MaxSessions:        2,
		ConcurrentSessions: 2,
		State:              models.Running,
	},
	{
		OtreeExperiment:    "xp_name",
		Namespace:          "fxt_live_sup_grouping",
		Slug:               "fxt_live_sup_grouping_slug",
		PerSession:         3,
		JoinOnce:           true,
		MaxSessions:        4,
		ConcurrentSessions: 1,
		SessionDuration:    sessionDurationTest,
		Grouping:           "What is your gender?\nMale:1\nFemale:1\nOther:1",
		StartedSessions:    0,
		State:              models.Running,
	},
	{
		OtreeExperiment:    "xp_name",
		Namespace:          "fxt_live_to_be_paused",
		Slug:               "fxt_live_to_be_paused_slug",
		PerSession:         4,
		MaxSessions:        2,
		ConcurrentSessions: 2,
		State:              models.Running,
	},
	{
		OtreeExperiment:    "xp_name",
		Namespace:          "fxt_live_sup_paused",
		Slug:               "fxt_live_sup_paused_slug",
		PerSession:         8,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		State:              models.Paused,
	},
	{
		OtreeExperiment:    "xp_name",
		Namespace:          "fxt_live_sup_busy",
		Slug:               "fxt_live_sup_busy_slug",
		PerSession:         2,
		MaxSessions:        2,
		ConcurrentSessions: 1,
		SessionDuration:    sessionDurationTest,
		State:              models.Running,
	},
	// participant tests
	{
		OtreeExperiment:    "xp_name",
		Namespace:          "fxt_live_ns1",
		Slug:               "fxt_live_ns1_slug",
		PerSession:         4,
		MaxSessions:        2,
		ConcurrentSessions: 2,
		State:              models.Running,
	},
	{
		OtreeExperiment:    "xp_name",
		Namespace:          "fxt_live_par_paused",
		Slug:               "fxt_live_par_paused_slug",
		PerSession:         8,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		State:              models.Paused,
	},
	{
		OtreeExperiment:    "xp_name",
		Namespace:          "fxt_live_par_launched",
		Slug:               "fxt_live_par_launched_slug",
		PerSession:         4,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		State:              models.Running,
	},
	{
		OtreeExperiment:    "xp_name",
		Namespace:          "fxt_live_par_completed",
		Slug:               "fxt_live_par_completed_slug",
		PerSession:         8,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		State:              models.Completed,
	},
	{
		OtreeExperiment:    "xp_name",
		Namespace:          "fxt_live_par_almost_completed",
		Slug:               "fxt_live_par_almost_completed_slug",
		PerSession:         4,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		StartedSessions:    3,
		State:              models.Running,
	},
	{
		OtreeExperiment:    "xp_name",
		Namespace:          "fxt_live_par_once",
		Slug:               "fxt_live_par_once_slug",
		PerSession:         4,
		JoinOnce:           true,
		MaxSessions:        2,
		ConcurrentSessions: 2,
		State:              models.Running,
	},
	{
		OtreeExperiment:    "xp_name",
		Namespace:          "fxt_live_par_redirect",
		Slug:               "fxt_live_par_redirect_slug",
		PerSession:         2,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		SessionDuration:    sessionDurationTest,
		StartedSessions:    0,
		State:              models.Running,
	},
	{
		OtreeExperiment:    "xp_name",
		Namespace:          "fxt_live_par_redirect2",
		Slug:               "fxt_live_par_redirect2_slug",
		PerSession:         2,
		JoinOnce:           true,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		SessionDuration:    sessionDurationTest,
		StartedSessions:    0,
		State:              models.Running,
	},
	{
		OtreeExperiment:    "xp_name",
		Namespace:          "fxt_live_par_reject",
		Slug:               "fxt_live_par_reject_slug",
		PerSession:         2,
		JoinOnce:           true,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		SessionDuration:    sessionDurationTest,
		StartedSessions:    0,
		State:              models.Running,
	},
	{
		OtreeExperiment:    "xp_name",
		Namespace:          "fxt_live_par_noreject",
		Slug:               "fxt_live_par_noreject_slug",
		PerSession:         2,
		JoinOnce:           false,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		SessionDuration:    sessionDurationTest,
		StartedSessions:    0,
		State:              models.Running,
	},
	// grouping tests
	{
		OtreeExperiment:    "xp_name",
		Namespace:          "fxt_live_grp",
		Slug:               "fxt_live_grp_slug",
		PerSession:         6,
		JoinOnce:           true,
		MaxSessions:        4,
		ConcurrentSessions: 1,
		SessionDuration:    sessionDurationTest,
		Grouping:           "What is your gender?\nMale:3\nFemale:3",
		StartedSessions:    0,
		State:              models.Running,
	},
}
