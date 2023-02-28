package router

import (
	"net/http"

	"github.com/ducksouplab/mastok/config"
	"github.com/gorilla/websocket"
)

// upgrader for websocket endpoints
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == config.OwnOrigin
	},
}

func reverse[T any](s []T) []T {
	a := make([]T, len(s))
	copy(a, s)

	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}

	return a
}
