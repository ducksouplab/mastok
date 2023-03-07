package router

import (
	"log"
	"net/http"
	"time"

	"github.com/ducksouplab/mastok/helpers"
	"github.com/ducksouplab/mastok/otree"
	"github.com/gin-gonic/gin"
)

type session struct {
	Id              string
	Code            string
	ConfigName      string  `json:"config_name"`
	CreatedAtFloat  float32 `json:"created_at"`
	NumParticipants int     `json:"num_participants"`
	AdminUrl        string  `json:"admin_url"`
}

type nestedSessionDetails struct {
	Config struct {
		Id string `json:"id"`
	} `json:"config"`
}

func (s session) FormatCreatedAt() string {
	return time.Unix(int64(s.CreatedAtFloat), 0).UTC().Format("2006-01-02 15:04:05")
}

func addSessionsRoutesTo(g *gin.RouterGroup) {
	g.GET("/sessions", func(c *gin.Context) {
		sessions := []session{}
		err := otree.GetOTreeJSON("/api/sessions", &sessions)

		if err != nil {
			log.Println(err)
			c.Status(http.StatusServiceUnavailable)
		} else {
			for i := range sessions { // use index to write to sessions
				code := sessions[i].Code
				sc := nestedSessionDetails{}
				err := otree.GetOTreeJSON("/api/sessions/"+code, &sc)
				if err != nil {
					log.Println(err)
					c.Status(http.StatusServiceUnavailable)
				}
				sessions[i].Id = sc.Config.Id
			}
			// reverse since oTree returns by chronogical create, we want latest first
			c.HTML(http.StatusOK, "sessions.tmpl", gin.H{
				"Sessions": helpers.Reverse(sessions),
			})
		}
	})
}
