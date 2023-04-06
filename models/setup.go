package models

import (
	"log"
	"os"
	"time"

	"github.com/ducksouplab/mastok/env"
	"github.com/ducksouplab/mastok/helpers"
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
	DB.AutoMigrate(&Campaign{}, &Session{}, &Participation{})
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
		consentString := helpers.ReadFile("consent.md")
		// dev fixtures
		var campaign = Campaign{
			OtreeExperiment:    "chatroulette",
			Namespace:          "dev_campaign_1",
			Slug:               "dev_campaign_1_slug",
			PerSession:         2,
			MaxSessions:        6,
			ConcurrentSessions: 1,
			Consent:            consentString,
			State:              Running,
		}
		// var session = Session{
		// 	Code:     "nztdjo76",
		// 	OtreeId:  "mk:dev_campaign_1:1",
		// 	Size:     4,
		// 	AdminUrl: "http://otree.host.com/SessionStartLinks/nztdjo76",
		// }
		if err := DB.Create(&campaign).Error; err != nil {
			log.Fatal(err)
		}
		// campaign.appendSession(&session)
		// simple campaign
		var otherCampaign = Campaign{
			OtreeExperiment:    "chatroulette",
			Namespace:          "dev_campaign_2",
			Slug:               "dev_campaign_2_slug",
			PerSession:         4,
			MaxSessions:        4,
			ConcurrentSessions: 1,
			Grouping:           "What is your gender?\nMale:2\nFemale:2\nChoose",
			SessionDuration:    2,
			Consent:            consentString,
			State:              Running,
		}
		if err := DB.Create(&otherCampaign).Error; err != nil {
			log.Fatal(err)
		}
	}
}
