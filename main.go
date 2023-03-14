package main

import (
	"log"

	"github.com/ducksouplab/mastok/env"
	"github.com/ducksouplab/mastok/front"
	"github.com/ducksouplab/mastok/models"
	"github.com/ducksouplab/mastok/router"
	"github.com/gin-gonic/gin"
)

func main() {
	if env.Mode == "DEV" || env.Mode == "BUILD_FRONT" {
		front.Build()
	} else if env.Mode == "RESET_DEV" {
		models.ReinitDevDB()
	}
	// command line mode, app stops
	if env.AsCommandLine {
		return
	}
	// main path
	models.ConnectAndMigrate()
	// HTTP server
	r := router.NewRouter(gin.Default())
	log.Println("[server] websocket origin allowed:", env.Origin)
	log.Println("[server] listening port:", env.Port)
	r.Run(":" + env.Port)
}
