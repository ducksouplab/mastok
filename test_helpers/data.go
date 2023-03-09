package test_helpers

import "github.com/ducksouplab/mastok/models"

// for DB
var FIXTURE_CAMPAIGNS []models.Campaign = []models.Campaign{
	{
		Namespace:        "fixture_ns1",
		Slug:             "fixture_slug1",
		ExperimentConfig: "config",
		PerSession:       4,
		SessionsMax:      2,
		State:            models.Running,
	},
	{
		Namespace:        "fixture_ns2_to_be_paused",
		Slug:             "fixture_slug2",
		ExperimentConfig: "config",
		PerSession:       4,
		SessionsMax:      2,
		State:            models.Running,
	},
	{
		Namespace:        "fixture_ns3_waiting",
		Slug:             "fixture_slug3",
		ExperimentConfig: "config",
		PerSession:       8,
		SessionsMax:      4,
		State:            models.Waiting,
	},
	{
		Namespace:        "fixture_ns4_waiting",
		Slug:             "fixture_slug4",
		ExperimentConfig: "config",
		PerSession:       8,
		SessionsMax:      4,
		State:            models.Waiting,
	},
}

// for oTree
type resource map[string]any

var SESSION_CONFIGS = []resource{
	{
		"real_world_currency_per_point": 1.0,
		"participation_fee":             0.0,
		"doc":                           "",
		"id":                            "CH",
		"name":                          "chatroulette",
		"display_name":                  "Chatroulette",
		"num_demo_participants":         10,
	},
	{
		"real_world_currency_per_point": 1.0,
		"participation_fee":             0.0,
		"doc":                           "",
		"id":                            "RA",
		"name":                          "rawroulette",
		"display_name":                  "Rawroulette",
		"num_demo_participants":         10,
	},
}
