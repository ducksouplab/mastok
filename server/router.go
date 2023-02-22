package server

import (
	"net/http"
	"path/filepath"

	"github.com/ducksouplab/mastok/helpers"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/multitemplate"
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

func createTemplateRenderer() multitemplate.Renderer {
	renderer := multitemplate.NewRenderer()

	includes, err := filepath.Glob(projectRoot + "server/templates/includes/*.tmpl")
	if err != nil {
		panic(err.Error())
	}

	for _, include := range includes {
		renderer.AddFromFiles(filepath.Base(include), projectRoot+"server/templates/layout.tmpl", include)
	}

	// first parameter is the exact name to be reused inside handler
	return renderer
}

func NewRouter() *gin.Engine {
	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.HTMLRender = createTemplateRenderer()
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		authBasicLogin: authBasicPassword,
	}))
	authorized.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/sessions")
	})
	addSessionsRoutes(authorized)

	return r
}
