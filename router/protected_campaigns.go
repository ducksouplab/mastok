package router

import (
	"log"
	"net/http"

	"github.com/ducksouplab/mastok/cache"
	"github.com/ducksouplab/mastok/env"
	"github.com/ducksouplab/mastok/helpers"
	"github.com/ducksouplab/mastok/live"
	"github.com/ducksouplab/mastok/models"
	"github.com/gin-gonic/gin"
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
			"Campaign":             model,
			"RenderedConsent":      helpers.MarkdownToHTML(model.Consent),
			"RenderedInstructions": helpers.MarkdownToHTML(model.Instructions),
			"RenderedPaused":       helpers.MarkdownToHTML(model.Paused),
			"RenderedCompleted":    helpers.MarkdownToHTML(model.Completed),
			"RenderedPending":      helpers.MarkdownToHTML(model.Pending),
		})
	})
	g.GET("/ws/campaigns/supervise", func(c *gin.Context) {
		wsSuperviseHandler(c.Writer, c.Request)
	})
	// CREATE
	g.GET("/campaigns/new", func(c *gin.Context) {
		c.HTML(http.StatusOK, "campaign_new.tmpl", gin.H{
			"Experiments": cache.GetOTreeConfigs(),
			"Campaign":    models.Campaign{},
		})
	})
	g.POST("/campaigns/new", func(c *gin.Context) {
		var input models.Campaign

		if err := c.ShouldBind(&input); err != nil {
			c.HTML(http.StatusUnprocessableEntity, "campaign_new.tmpl", gin.H{
				"Experiments": cache.GetOTreeConfigs(),
				"Error":       changeErrorMessage(err.Error()),
				"Campaign":    input,
			})
			return
		}

		if err := models.DB.Create(&input).Error; err != nil {
			c.HTML(http.StatusUnprocessableEntity, "campaign_new.tmpl", gin.H{
				"Experiments": cache.GetOTreeConfigs(),
				"Error":       changeErrorMessage(err.Error()),
				"Campaign":    input,
			})
			return
		}

		c.Redirect(http.StatusFound, env.WebPrefix+"/campaigns")
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
		input.OTreeConfigName = model.OTreeConfigName

		if err := c.ShouldBind(&input); err != nil {
			c.HTML(http.StatusUnprocessableEntity, "campaign_edit.tmpl", gin.H{
				"Error":    changeErrorMessage(err.Error()),
				"Campaign": input,
			})
			return
		}
		// pick fields to update:
		// - exclude Namespace and OTreeConfigName
		// - and force zero values updates for Grouping and Consent
		selecteds := []string{"Slug", "PerSession", "JoinOnce", "MaxSessions", "ConcurrentSessions", "SessionDuration", "WaitingLimit", "Grouping", "Consent", "Instructions", "Paused", "Completed", "Pending"}
		if err := models.DB.Model(&model).Select(selecteds).Updates(input).Error; err != nil {
			c.HTML(http.StatusUnprocessableEntity, "campaign_edit.tmpl", gin.H{
				"Error":    changeErrorMessage(err.Error()),
				"Campaign": input,
			})
			return
		} else {
			// it would be better to do this in an AfterUpdate hook, but it introduces an import cycle
			live.UpdateRunner(model)
		}

		c.Redirect(http.StatusFound, env.WebPrefix+"/campaigns/supervise/"+namespace)
	})
}
