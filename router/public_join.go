package router

import (
	"crypto/sha512"
	"encoding/hex"
	"log"
	"net/http"

	"github.com/ducksouplab/mastok/live"
	"github.com/ducksouplab/mastok/models"
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
		slug := c.Param("slug")
		model, ok := models.GetCampaignBySlug(slug)
		if !ok {
			log.Printf("[router] find campaign failed for namespace %v", slug)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.HTML(http.StatusOK, "join.tmpl", gin.H{
			"Campaign": model,
			"Slug":     c.Param("slug"),
			"Hash":     getClientHash(c),
		})
	})
	g.GET("/ws/join", func(c *gin.Context) {
		wsJoinHandler(c.Writer, c.Request)
	})
}

func getClientHash(c *gin.Context) string {
	clientInfo := c.ClientIP() + c.RemoteIP()
	for _, header := range []string{"Accept", "Accept-Encoding", "Accept-Language", "Sec-Ch-Ua", "Sec-Ch-Ua-Mobile", "Sec-Ch-Ua-Platform"} {
		for _, h := range c.Request.Header[header] {
			clientInfo += h
		}
	}
	hash := sha512.Sum512([]byte(clientInfo))
	return hex.EncodeToString(hash[:])
}
