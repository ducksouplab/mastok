package router

import (
	"log"
	"os"
	"testing"

	"github.com/ducksouplab/mastok/models"
	"github.com/gin-gonic/gin"
)

var tr *gin.Engine

func getTestRouter() *gin.Engine {
	return tr
}

func TestMain(m *testing.M) {
	tr = NewRouter(gin.New())
	models.ReinitTestDB()
	if err := models.DB.Create(Fixtures).Error; err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
}

var Fixtures []models.Campaign = []models.Campaign{
	{
		OtreeExperiment:    "xp_name",
		Namespace:          "fxt_router_ns1",
		Slug:               "fxt_router_ns1_slug",
		PerSession:         4,
		MaxSessions:        2,
		ConcurrentSessions: 2,
		Consent:            "#Title\ntext\n[accept]Accept[/accept]",
		State:              models.Running,
	},
	{
		OtreeExperiment:    "xp_name",
		Namespace:          "fxt_router_ns2_edit",
		Slug:               "fxt_router_ns2_edit_slug",
		PerSession:         4,
		MaxSessions:        2,
		ConcurrentSessions: 2,
		Consent:            "#Title\ntext\n[accept]Accept[/accept]",
		State:              models.Running,
	},
}
