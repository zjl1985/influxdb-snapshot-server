package router

import (
	. "fastdb-server/controller"
	"fastdb-server/models"
	"github.com/gin-gonic/gin"
	"io"
	"os"
)

func InitRouter(config *models.Config) *gin.Engine {
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)
	gin.SetMode(config.Mode)
	router := gin.Default()
	initialize(config)
	router.POST("/snapshot", Snapshot)
	router.POST("/snapshot/write", InfluxSub)
	return router
}

func initialize(config *models.Config) {
	InitSnapshot(config.Delay)
}
