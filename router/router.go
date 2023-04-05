package router

import (
	"html/template"
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

	withLayouts, err := filepath.Glob(env.ProjectRoot + "templates/with_layout/*.tmpl")
	if err != nil {
		panic(err.Error())
	}

	for _, t := range withLayouts {
		renderer.AddFromFilesFuncs(
			filepath.Base(t),
			template.FuncMap{
				"WebPrefix": func() string { return env.WebPrefix },
			},
			env.ProjectRoot+"templates/layout.tmpl",
			t,
		)
	}

	withoutLayouts, err := filepath.Glob(env.ProjectRoot + "templates/without_layout/*.tmpl")
	if err != nil {
		panic(err.Error())
	}

	for _, t := range withoutLayouts {
		renderer.AddFromFilesFuncs(
			filepath.Base(t),
			template.FuncMap{
				"WebPrefix": func() string { return env.WebPrefix },
			},
			t,
		)
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
	addCustomValidators()

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
		c.Redirect(http.StatusFound, env.WebPrefix+"/dashboard")
	})
	// add routes
	addDashboardRoutesTo(authorizedGroup)
	addCampaignsRoutesTo(authorizedGroup)
	addSessionsRoutesTo(authorizedGroup)

	return r
}
