package main

import (
	"log"

	"github.com/ducksouplab/mastok/config"
	"github.com/ducksouplab/mastok/front"
	"github.com/ducksouplab/mastok/models"
	"github.com/ducksouplab/mastok/router"
)

func main() {
	if config.OwnEnv == "DEV" || config.OwnEnv == "BUILD_FRONT" {
		front.Build()
	}
	// command line mode...
	if config.OwnEnv == "BUILD_FRONT" {
		return
	}
	// ...or server mode, starts with DB
	models.ConnectAndMigrate()
	// then HTTP server
	r := router.NewRouter()
	log.Println("[server] websocket origin allowed:", config.OwnOrigin)
	log.Println("[server] listening port:", config.OwnPort)
	r.Run(":" + config.OwnPort)
}
