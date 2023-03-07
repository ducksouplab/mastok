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
	}
	// command line mode...
	if env.Mode == "BUILD_FRONT" {
		return
	}
	// ...or server mode, starts with DB
	models.ConnectAndMigrate()
	// then HTTP server
	r := router.NewRouter(gin.Default())
	log.Println("[server] websocket origin allowed:", env.Origin)
	log.Println("[server] listening port:", env.Port)
	r.Run(":" + env.Port)
}
