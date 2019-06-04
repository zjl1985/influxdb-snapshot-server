package router

import (
	"fastdb-server/controller"
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
	router.POST("/snapshot", controller.Snapshot)
	router.POST("/snapshot/write", controller.InfluxSub)

	router.GET("/tags/:database", controller.SelectPage)
	router.GET("/tags/:database/:code", controller.SelectPage)
	router.GET("/tag/:id", controller.SelectById)
	router.POST("/tag", controller.Create)
	router.POST("/tags", controller.CreateList)
	router.PUT("/tag", controller.Update)
	router.DELETE("/tag/:id", controller.Delete)
	router.DELETE("/tags", controller.DeleteList)
	return router
}

func initialize(config *models.Config) {
	controller.InitSnapshot(config.Delay)
}
