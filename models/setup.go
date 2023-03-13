package models

import (
	"os"

	"github.com/ducksouplab/mastok/env"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

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
