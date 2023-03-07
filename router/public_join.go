package router

import (
	"log"
	"net/http"

	"github.com/ducksouplab/mastok/live"
	"github.com/gin-gonic/gin"
)

func wsJoinHandler(w http.ResponseWriter, r *http.Request) {
	// upgrade HTTP request to Websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("[router] supervise websocket upgrade failed")
		return
	}
	log.Println("[router] supervise websocket upgrade success")

	live.RunParticipant(ws, r.FormValue("slug"))
}

func addJoinRoutesTo(g *gin.RouterGroup) {
	g.GET("/join/:slug", func(c *gin.Context) {
		c.HTML(http.StatusOK, "join.tmpl", gin.H{
			"Slug": c.Param("slug"),
		})
	})
	g.GET("/ws/join", func(c *gin.Context) {
		wsJoinHandler(c.Writer, c.Request)
	})
}
