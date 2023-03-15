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
		Namespace:          "fxt_router_ns1",
		Slug:               "fxt_router_ns1_slug",
		OtreeExperiment:    "xp_name",
		PerSession:         4,
		MaxSessions:        2,
		ConcurrentSessions: 2,
		State:              models.Running,
	},
}
