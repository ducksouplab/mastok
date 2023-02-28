package models

import (
	"github.com/ducksouplab/mastok/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectAndMigrate() {
	var err error

	if config.OwnEnv == "TEST" {
		DB, err = gorm.Open(sqlite.Open(config.OwnRoot+"test.db"), &gorm.Config{})
	} else if config.OwnEnv == "DEV" || config.DatabaseURL == "" {
		DB, err = gorm.Open(sqlite.Open(config.OwnRoot+"local.db"), &gorm.Config{})
	} else {
		DB, err = gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{})
	}
	if err != nil {
		panic("failed to connect database")
	}
	// Migrate the schema
	DB.AutoMigrate(&Campaign{})
}
