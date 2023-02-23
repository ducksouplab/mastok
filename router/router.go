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

	includes, err := filepath.Glob(config.ProjectRoot + "templates/includes/*.tmpl")
	if err != nil {
		panic(err.Error())
	}

	for _, include := range includes {
		renderer.AddFromFiles(filepath.Base(include), config.ProjectRoot+"templates/layout.tmpl", include)
	}

	// first parameter is the exact name to be reused inside handler
	return renderer
}

func NewRouter() *gin.Engine {
	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.HTMLRender = createTemplateRenderer()

	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		config.AuthBasicLogin: config.AuthBasicPassword,
	}))
	authorized.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/dashboard")
	})

	addDashboardRoutesTo(authorized)
	addCampaignsRoutesTo(authorized)
	addSessionsRoutesTo(authorized)

	return r
}
