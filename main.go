package main

import (
	"github.com/ducksouplab/mastok/config"
	"github.com/ducksouplab/mastok/models"
	"github.com/ducksouplab/mastok/router"
)

func main() {
	// DB
	models.ConnectAndMigrate()
	// HTTP server
	r := router.NewRouter()
	r.Run(":" + config.OwnPort)
}
