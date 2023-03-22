package router

import (
	"log"
	"net/http"

	"github.com/ducksouplab/mastok/cache"
	"github.com/ducksouplab/mastok/live"
	"github.com/ducksouplab/mastok/models"
	"github.com/gin-gonic/gin"
	"github.com/shurcooL/github_flavored_markdown"
)

func wsSuperviseHandler(w http.ResponseWriter, r *http.Request) {
	// upgrade HTTP request to Websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("[router] supervise websocket upgrade failed")
		return
	}
	log.Println("[router] supervise websocket upgrade success")

	live.RunSupervisor(ws, r.FormValue("namespace"))
}

func addCampaignsRoutesTo(g *gin.RouterGroup) {
	g.GET("/campaigns", func(c *gin.Context) {
		var campaigns []models.Campaign
		if err := models.DB.Order("ID desc").Find(&campaigns).Error; err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.HTML(http.StatusOK, "campaigns.tmpl", gin.H{
			"Campaigns": campaigns,
		})
	})
	g.GET("/campaigns/new", func(c *gin.Context) {
		c.HTML(http.StatusOK, "campaigns_new.tmpl", gin.H{
			"Experiments": cache.GetExperiments(),
		})
	})
	g.GET("/campaigns/supervise/:namespace", func(c *gin.Context) {
		namespace := c.Param("namespace")
		campaign, ok := models.GetCampaignByNamespace(namespace)
		if !ok {
			log.Printf("[router] find campaign failed for namespace %v", namespace)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.HTML(http.StatusOK, "campaigns_supervise.tmpl", gin.H{
			"Campaign":        campaign,
			"RenderedConsent": string(github_flavored_markdown.Markdown([]byte("input:\n\n- [ ] item"))),
		})
	})
	g.GET("/ws/campaigns/supervise", func(c *gin.Context) {
		wsSuperviseHandler(c.Writer, c.Request)
	})
	g.POST("/campaigns", func(c *gin.Context) {
		var campaign models.Campaign

		if err := c.ShouldBind(&campaign); err != nil {
			c.HTML(http.StatusUnprocessableEntity, "campaigns_new.tmpl", gin.H{
				"Experiments": cache.GetExperiments(),
				"Error":       err.Error(),
			})
			return
		}

		if err := models.DB.Create(&campaign).Error; err != nil {
			c.HTML(http.StatusUnprocessableEntity, "campaigns_new.tmpl", gin.H{
				"Experiments": cache.GetExperiments(),
				"Error":       err.Error(),
			})
			return
		}

		c.Redirect(http.StatusFound, "/campaigns")
	})
}
