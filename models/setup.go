package models

import (
	"log"
	"os"
	"time"

	"github.com/ducksouplab/mastok/env"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB
var SessionDurationUnit time.Duration

func init() {
	if env.Mode == "TEST" {
		SessionDurationUnit = 3 * time.Millisecond
	} else {
		SessionDurationUnit = time.Minute
	}
}

func ConnectAndMigrate() {
	var err error

	if env.Mode == "TEST" {
		DB, err = gorm.Open(sqlite.Open(env.ProjectRoot+"test.db"), &gorm.Config{})
	} else if env.Mode == "DEV" || env.DatabaseURL == "" {
		DB, err = gorm.Open(sqlite.Open(env.ProjectRoot+"local.db"), &gorm.Config{})
	} else {
		DB, err = gorm.Open(postgres.Open(env.DatabaseURL), &gorm.Config{})
	}
	if err != nil {
		panic("failed to connect database")
	}
	// Migrate the schema
	DB.AutoMigrate(&Campaign{}, &Session{})
}

// not declared in a _test.go file to be callable from another test package
func ReinitTestDB() {
	if env.Mode == "TEST" {
		os.Remove(env.ProjectRoot + "test.db")
		ConnectAndMigrate()
	}
}

func ReinitDevDB() {
	if env.Mode == "RESET_DEV" {
		os.Remove(env.ProjectRoot + "local.db")
		ConnectAndMigrate()
		// dev fixtures
		var campaign = Campaign{
			Namespace:          "dev_campaign_1",
			Slug:               "dev_campaign_1_slug",
			OtreeExperimentId:  "CH",
			PerSession:         4,
			MaxSessions:        2,
			ConcurrentSessions: 1,
			State:              Running,
		}
		var session = Session{
			Code:     "nztdjo76",
			OtreeId:  "mk:dev_campaign_1:1",
			Size:     4,
			AdminUrl: "http://localhost:8180/SessionStartLinks/nztdjo76",
		}
		if err := DB.Create(&campaign).Error; err != nil {
			log.Fatal(err)
		}
		campaign.appendSession(session)
		// simple campaign
		var otherCampaign = Campaign{
			Namespace:          "dev_campaign_2",
			Slug:               "dev_campaign_2_slug",
			OtreeExperimentId:  "CH",
			PerSession:         2,
			MaxSessions:        4,
			SessionDuration:    2,
			ConcurrentSessions: 1,
			State:              Running,
		}
		if err := DB.Create(&otherCampaign).Error; err != nil {
			log.Fatal(err)
		}
	}
}
