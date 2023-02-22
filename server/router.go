package server

import (
	"net/http"

	"github.com/ducksouplab/mastok/helpers"
	"github.com/gin-gonic/gin"
)

var projectRoot, authBasicLogin, authBasicPassword, otreeUrl, otreeRestKey string

func init() {
	// test are not executed at the project root, so using LoadHTMLGlob needs absolute path
	projectRoot = helpers.GetenvOr("MASTOK_PROJECT_ROOT", ".") + "/"
	authBasicLogin = helpers.GetenvOr("MASTOK_LOGIN", "admin")
	authBasicPassword = helpers.GetenvOr("MASTOK_PASSWORD", "admin")
	otreeUrl = helpers.GetenvOr("MASTOK_OTREE_URL", "http://localhost:8180")
	otreeRestKey = helpers.GetenvOr("MASTOK_OTREE_REST_KEY", "key")
}

func NewRouter() *gin.Engine {
	r := gin.Default()
	r.LoadHTMLGlob(projectRoot + "server/templates/*")
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		authBasicLogin: authBasicPassword,
	}))
	authorized.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/sessions")
	})
	addSessionsRoutes(authorized)

	return r
}
