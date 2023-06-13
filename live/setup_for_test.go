package live

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/ducksouplab/mastok/models"
)

const (
	sessionDurationTest = 60
	shortDuration       = 10 * time.Millisecond
	longDuration        = 100 * time.Millisecond // for instance if there are DB writes
	longerDuration      = 300 * time.Millisecond // more DB writes
)

func TestMain(m *testing.M) {
	models.ReinitTestDB()
	if err := models.DB.Create(Fixtures).Error; err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
}

var Fixtures []models.Campaign = []models.Campaign{
	// runner tests
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_run",
		Slug:               "fxt_run_slug",
		PerSession:         4,
		MaxSessions:        2,
		ConcurrentSessions: 2,
		State:              models.Running,
	},
	// otree tests
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_otree_to_be_launched",
		Slug:               "fxt_otree_to_be_launched_slug",
		PerSession:         4,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		State:              models.Running,
	},
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_otree_groups_to_be_launched",
		Slug:               "fxt_otree_groups_to_be_launched_slug",
		PerSession:         4,
		JoinOnce:           true,
		MaxSessions:        4,
		ConcurrentSessions: 1,
		SessionDuration:    sessionDurationTest,
		Grouping:           "What is your gender?\nMale:2\nFemale:2\nChoose",
		StartedSessions:    0,
		State:              models.Running,
	},
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_otree_busy",
		Slug:               "fxt_otree_busy_slug",
		PerSession:         5,
		MaxSessions:        3,
		ConcurrentSessions: 1,
		SessionDuration:    100 * sessionDurationTest, // important for test
		StartedSessions:    0,
		State:              models.Running,
	},
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_otree_concurrent",
		Slug:               "fxt_otree_concurrent_slug",
		PerSession:         2,
		MaxSessions:        10,
		ConcurrentSessions: 2,
		SessionDuration:    sessionDurationTest,
		StartedSessions:    0,
		State:              models.Running,
	},
	// connect tests
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_connect_grouping",
		Slug:               "fxt_connect_grouping_slug",
		PerSession:         4,
		MaxSessions:        4,
		ConcurrentSessions: 1,
		Grouping:           "What is your gender?\nMale:2\nFemale:2\nChoose",
		StartedSessions:    0,
		State:              models.Running,
	},
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_connect_pending_busy",
		Slug:               "fxt_connect_pending_busy_slug",
		PerSession:         4,
		JoinOnce:           true,
		MaxSessions:        4,
		ConcurrentSessions: 1,
		SessionDuration:    100 * sessionDurationTest, // important for test
		Grouping:           "What is your gender?\nMale:2\nFemale:2\nChoose",
		StartedSessions:    0,
		State:              models.Running,
	},
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_par_pending",
		Slug:               "fxt_par_pending_slug",
		PerSession:         4,
		MaxSessions:        3,
		ConcurrentSessions: 2,
		StartedSessions:    0,
		State:              models.Busy, // important for test
	},
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_par_connect_full",
		Slug:               "fxt_par_connect_full_slug",
		PerSession:         4,
		MaxSessions:        3,
		ConcurrentSessions: 2,
		StartedSessions:    1,
		State:              models.Busy, // important for test
	},
	// supervisor tests
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_sup",
		Slug:               "fxt_sup_slug",
		PerSession:         4,
		MaxSessions:        2,
		ConcurrentSessions: 2,
		State:              models.Running,
	},
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_sup_grouping",
		Slug:               "fxt_sup_grouping_slug",
		PerSession:         3,
		JoinOnce:           true,
		MaxSessions:        4,
		ConcurrentSessions: 1,
		SessionDuration:    sessionDurationTest,
		Grouping:           "What is your gender?\nMale:1\nFemale:1\nOther:1\nChoose",
		StartedSessions:    0,
		State:              models.Running,
	},
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_to_be_paused",
		Slug:               "fxt_to_be_paused_slug",
		PerSession:         4,
		MaxSessions:        2,
		ConcurrentSessions: 2,
		State:              models.Running,
	},
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_sup_paused",
		Slug:               "fxt_sup_paused_slug",
		PerSession:         8,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		State:              models.Paused,
	},
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_sup_paused2",
		Slug:               "fxt_sup_paused2_slug",
		PerSession:         8,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		State:              models.Paused,
	},
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_sup_busy",
		Slug:               "fxt_sup_busy_slug",
		PerSession:         2,
		MaxSessions:        2,
		ConcurrentSessions: 1,
		SessionDuration:    sessionDurationTest,
		State:              models.Running,
	},
	// participant tests
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_par",
		Slug:               "fxt_par_slug",
		PerSession:         4,
		MaxSessions:        2,
		ConcurrentSessions: 2,
		State:              models.Running,
	},
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_par_paused",
		Slug:               "fxt_par_paused_slug",
		PerSession:         8,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		State:              models.Paused,
	},
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_par_completed",
		Slug:               "fxt_par_completed_slug",
		PerSession:         8,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		State:              models.Completed,
	},
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_par_almost_completed",
		Slug:               "fxt_par_almost_completed_slug",
		PerSession:         4,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		StartedSessions:    3,
		State:              models.Running,
	},
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_par_once",
		Slug:               "fxt_par_once_slug",
		PerSession:         4,
		JoinOnce:           true,
		MaxSessions:        2,
		ConcurrentSessions: 2,
		State:              models.Running,
	},
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_par_redirect",
		Slug:               "fxt_par_redirect_slug",
		PerSession:         2,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		SessionDuration:    sessionDurationTest,
		StartedSessions:    0,
		State:              models.Running,
	},
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_par_redirect2",
		Slug:               "fxt_par_redirect2_slug",
		PerSession:         2,
		JoinOnce:           true,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		SessionDuration:    sessionDurationTest,
		StartedSessions:    0,
		State:              models.Running,
	},
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_par_reject",
		Slug:               "fxt_par_reject_slug",
		PerSession:         2,
		JoinOnce:           true,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		SessionDuration:    sessionDurationTest,
		StartedSessions:    0,
		State:              models.Running,
	},
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_par_noreject",
		Slug:               "fxt_par_noreject_slug",
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
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_grp",
		Slug:               "fxt_grp_slug",
		PerSession:         6,
		JoinOnce:           true,
		MaxSessions:        4,
		ConcurrentSessions: 1,
		SessionDuration:    sessionDurationTest,
		Grouping:           "What is your gender?\nMale:3\nFemale:3\nChoose",
		StartedSessions:    0,
		State:              models.Running,
	},
	// Instruction tests
	{
		OTreeConfigName:    "test_config_1_to_8",
		Namespace:          "fxt_instructions",
		Slug:               "fxt_instructions_slug",
		PerSession:         2,
		JoinOnce:           true,
		MaxSessions:        4,
		ConcurrentSessions: 1,
		SessionDuration:    sessionDurationTest,
		Instructions:       "#Title\nParagraph\n",
		StartedSessions:    0,
		State:              models.Running,
	},
}
