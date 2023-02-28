package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func addDashboardRoutesTo(g *gin.RouterGroup) {
	g.GET("/dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dashboard.tmpl", nil)
	})
}
