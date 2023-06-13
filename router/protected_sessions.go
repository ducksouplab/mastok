package router

import (
	"log"
	"net/http"

	"github.com/ducksouplab/mastok/helpers"
	"github.com/ducksouplab/mastok/otree"
	"github.com/gin-gonic/gin"
)

const DISPLAYED_SESSIONS = 25

func addSessionsRoutesTo(g *gin.RouterGroup) {
	g.GET("/sessions", func(c *gin.Context) {
		sessions := []otree.Session{}
		err := otree.GetOTreeJSON("/api/sessions", &sessions)

		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}
		// most recent first
		sessions = helpers.Reverse(sessions)
		// limit HTTP requests
		if len(sessions) > DISPLAYED_SESSIONS {
			sessions = sessions[:DISPLAYED_SESSIONS]
		}

		for i := range sessions { // use index to write to sessions
			code := sessions[i].Code
			sc := otree.NestedSessionDetails{}

			err := otree.GetOTreeJSON("/api/sessions/"+code, &sc)
			if err != nil {
				log.Println(err)
				c.AbortWithStatus(http.StatusServiceUnavailable)
				return
			}
			sessions[i].Id = sc.Config.Id

		}
		// reverse since oTree returns by chronogical create, we want latest first
		c.HTML(http.StatusOK, "sessions.tmpl", gin.H{
			"Sessions": sessions,
		})
	})
}
