package models

import (
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	ReinitTestDB()
	if err := DB.Create(Fixtures).Error; err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
}

var Fixtures []Campaign = []Campaign{
	{
		Namespace:          "fxt_models_ns1",
		Slug:               "fxt_models_ns1_slug",
		Config:             "config",
		PerSession:         4,
		MaxSessions:        4,
		ConcurrentSessions: 2,
		StartedSessions:    3,
		State:              Running,
	},
}
