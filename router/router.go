package router

import (
	"net/http"
	"path/filepath"

	"github.com/ducksouplab/mastok/env"
	"github.com/ducksouplab/mastok/helpers"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func createTemplateRenderer() multitemplate.Renderer {
	renderer := multitemplate.NewRenderer()

	includes, err := filepath.Glob(env.ProjectRoot + "templates/includes/*.tmpl")
	if err != nil {
		panic(err.Error())
	}

	for _, include := range includes {
		renderer.AddFromFiles(filepath.Base(include), env.ProjectRoot+"templates/layout.tmpl", include)
	}

	// first parameter is the exact name to be reused inside handler
	return renderer
}

// upgrader for websocket endpoints
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return helpers.Contains(env.AllowedOrigins, origin)
	},
}

func NewRouter(r *gin.Engine) *gin.Engine {
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.HTMLRender = createTemplateRenderer()

	// static assets
	r.Static(env.WebPrefix+"/assets", "./front/static/assets")
	// public routes
	publicGroup := r.Group(env.WebPrefix)
	addJoinRoutesTo(publicGroup)
	// protect routes
	authorizedGroup := r.Group(env.WebPrefix, gin.BasicAuth(gin.Accounts{
		env.BasicLogin: env.BasicPassword,
	}))
	authorizedGroup.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/dashboard")
	})
	// add routes
	addDashboardRoutesTo(authorizedGroup)
	addCampaignsRoutesTo(authorizedGroup)
	addSessionsRoutesTo(authorizedGroup)

	return r
}
