package router

import (
	"net/http"

	"github.com/ducksouplab/mastok/models"
	"github.com/ducksouplab/mastok/otree"
	"github.com/gin-gonic/gin"
)

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
			"Experiments": otree.GetExperimentCache(),
		})
	})
	g.POST("/campaigns", func(c *gin.Context) {
		var campaign models.Campaign
		if err := c.ShouldBind(&campaign); err != nil {
			c.HTML(http.StatusUnprocessableEntity, "campaigns_new.tmpl", gin.H{
				"Experiments": otree.GetExperimentCache(),
				"Error":       err.Error(),
			})
			return
		}

		if err := models.DB.Create(&campaign).Error; err != nil {
			c.HTML(http.StatusUnprocessableEntity, "campaigns_new.tmpl", gin.H{
				"Experiments": otree.GetExperimentCache(),
				"Error":       err.Error(),
			})
			return
		}

		c.Redirect(http.StatusFound, "/campaigns")
	})
}
