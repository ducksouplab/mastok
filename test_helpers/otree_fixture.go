package test_helpers

import "github.com/ducksouplab/mastok/env"

type resource map[string]any

var OTREE_GET_SESSION_CONFIGS = []resource{
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

// created with: session_config_name='chatroulette' num_participants:=4 modified_session_config_fields:='{"id": "mk:namespace"}'
var OTREE_POST_SESSION = resource{
	"admin_url":        env.OTreeURL + "/SessionStartLinks/5c7drkqy",
	"code":             "5c7drkqy",
	"session_wide_url": env.OTreeURL + "/join/nubogeke",
}

var OTREE_GET_SESSION = resource{
	"code":             "5c7drkqy",
	"num_participants": 4,
	"created_at":       1678359821.3008485,
	"label":            "",
	"config_name":      "chatroulette",
	"config": resource{
		"real_world_currency_per_point": 1.0,
		"participation_fee":             0.0,
		"doc":                           "",
		"id":                            "mk:namespace:1",
		"name":                          "chatroulette",
		"display_name":                  "Chatroulette",
		"app_sequence":                  []string{"chatroulette"},
		"num_demo_participants":         10,
	},
	"REAL_WORLD_CURRENCY_CODE": "EUR",
	"session_wide_url":         "http://localhost:8180/join/fonatoje",
	"admin_url":                "http://localhost:8180/SessionStartLinks/t1wlmb4v",
	"participants": []resource{
		{
			"id_in_session":                 1,
			"code":                          "vf6xq8fx",
			"label":                         nil,
			"payoff_in_real_world_currency": 0.0,
		},
		{
			"id_in_session":                 2,
			"code":                          "55ld3hp4",
			"label":                         nil,
			"payoff_in_real_world_currency": 0.0,
		},
		{
			"id_in_session":                 3,
			"code":                          "c1a73h31",
			"label":                         nil,
			"payoff_in_real_world_currency": 0.0,
		},
		{
			"id_in_session":                 4,
			"code":                          "8h9t7wxl",
			"label":                         nil,
			"payoff_in_real_world_currency": 0.0,
		},
	},
}