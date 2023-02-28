package router

import (
	"net/http"
	"path/filepath"

	"github.com/ducksouplab/mastok/config"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
)

func createTemplateRenderer() multitemplate.Renderer {
	renderer := multitemplate.NewRenderer()

	includes, err := filepath.Glob(config.OwnRoot + "templates/includes/*.tmpl")
	if err != nil {
		panic(err.Error())
	}

	for _, include := range includes {
		renderer.AddFromFiles(filepath.Base(include), config.OwnRoot+"templates/layout.tmpl", include)
	}

	// first parameter is the exact name to be reused inside handler
	return renderer
}

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gzip.Gzip(gzip.DefaultCompression), gin.Recovery())
	r.HTMLRender = createTemplateRenderer()

	// static assets
	r.Static(config.OwnWebPrefix+"/assets", "./front/static/assets")
	// public routes
	publicGroup := r.Group(config.OwnWebPrefix)
	addJoinRoutesTo(publicGroup)
	// protect routes
	authorizedGroup := r.Group(config.OwnWebPrefix, gin.BasicAuth(gin.Accounts{
		config.OwnBasicLogin: config.OwnBasicPassword,
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
