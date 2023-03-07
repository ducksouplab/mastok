package router

import (
	"github.com/gin-gonic/gin"
)

var tr *gin.Engine

func init() {
	// don't use gin.Default() request login and recovery
	tr = NewRouter(gin.New())
}

func getTestRouter() *gin.Engine {
	return tr
}
