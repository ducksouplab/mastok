package router

import (
	"html/template"
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
	// list
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
	// supervise
	g.GET("/campaigns/supervise/:namespace", func(c *gin.Context) {
		namespace := c.Param("namespace")
		model, ok := models.GetCampaignByNamespace(namespace)
		if !ok {
			log.Printf("[router] find campaign failed for namespace %v", namespace)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.HTML(http.StatusOK, "campaign_supervise.tmpl", gin.H{
			"Campaign":        model,
			"RenderedConsent": template.HTML(github_flavored_markdown.Markdown([]byte(model.Consent))),
		})
	})
	g.GET("/ws/campaigns/supervise", func(c *gin.Context) {
		wsSuperviseHandler(c.Writer, c.Request)
	})
	// CREATE
	g.GET("/campaigns/new", func(c *gin.Context) {
		c.HTML(http.StatusOK, "campaign_new.tmpl", gin.H{
			"Experiments": cache.GetExperiments(),
			"Campaign":    models.Campaign{},
		})
	})
	g.POST("/campaigns/new", func(c *gin.Context) {
		var input models.Campaign

		if err := c.ShouldBind(&input); err != nil {
			c.HTML(http.StatusUnprocessableEntity, "campaign_new.tmpl", gin.H{
				"Experiments": cache.GetExperiments(),
				"Error":       err.Error(),
				"Campaign":    input,
			})
			return
		}

		if err := models.DB.Create(&input).Error; err != nil {
			c.HTML(http.StatusUnprocessableEntity, "campaign_new.tmpl", gin.H{
				"Experiments": cache.GetExperiments(),
				"Error":       err.Error(),
				"Campaign":    input,
			})
			return
		}

		c.Redirect(http.StatusFound, "/campaigns")
	})
	// EDIT
	g.GET("/campaigns/edit/:namespace", func(c *gin.Context) {
		namespace := c.Param("namespace")
		model, ok := models.GetCampaignByNamespace(namespace)
		if !ok {
			log.Printf("[router] find campaign failed for namespace %v", namespace)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.HTML(http.StatusOK, "campaign_edit.tmpl", gin.H{
			"Campaign": model,
		})
	})
	g.POST("/campaigns/edit/:namespace", func(c *gin.Context) {
		namespace := c.Param("namespace")
		model, ok := models.GetCampaignByNamespace(namespace)
		if !ok {
			log.Printf("[router] find campaign failed for namespace %v", namespace)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		var input models.Campaign
		input.Namespace = model.Namespace
		input.OtreeExperiment = model.OtreeExperiment

		if err := c.ShouldBind(&input); err != nil {
			c.HTML(http.StatusUnprocessableEntity, "campaign_edit.tmpl", gin.H{
				"Error":    err.Error(),
				"Campaign": input,
			})
			return
		}
		// pick fileds to update:
		// - exclude Namespace and OtreeExperiment
		// - and force zero values updates for Grouping and Consent
		selecteds := []string{"Slug", "PerSession", "JoinOnce", "MaxSessions", "ConcurrentSessions", "SessionDuration", "WaitingLimit", "Grouping", "Consent"}
		if err := models.DB.Model(&model).Select(selecteds).Updates(input).Error; err != nil {
			c.HTML(http.StatusUnprocessableEntity, "campaign_edit.tmpl", gin.H{
				"Error":    err.Error(),
				"Campaign": input,
			})
			return
		}

		c.Redirect(http.StatusFound, "/campaigns/supervise/"+namespace)
	})
}
