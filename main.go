package main

import (
	"github.com/ducksouplab/mastok/helpers"
	"github.com/ducksouplab/mastok/server"
)

var port string

func init() {
	port = helpers.GetenvOr("MASTOK_PORT", "8190")
}

func main() {
	r := server.NewRouter()
	r.Run(":" + port)
}
