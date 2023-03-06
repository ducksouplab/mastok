package router

import (
	"fmt"
	"net/http"
	"os"

	"github.com/ducksouplab/mastok/env"
	"github.com/ducksouplab/mastok/helpers"
	"github.com/gorilla/websocket"
)

// upgrader for websocket endpoints
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		fmt.Fprintf(os.Stdout, ">>>>>>>>>>>>>>>>>>>>>>>>>> %+v\n", r.Header)
		return helpers.Contains(env.AllowedOrigins, origin)
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
