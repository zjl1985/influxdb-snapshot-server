package router

import (
	"fastdb-server/controller"
	"fastdb-server/controller/database"
	"fastdb-server/controller/influx"
	"fastdb-server/models"
	"github.com/gin-gonic/gin"
	"io"
	"os"
)

// InitRouter 注册服务
func InitRouter(config *models.Config) *gin.Engine {
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)
	gin.SetMode(config.Mode)
	router := gin.Default()
	initialize(config)
	api := router.Group("/api")
	api.POST("/snapshot", controller.Snapshot)
	api.POST("/snapshot/write", controller.InfluxSub)

	api.GET("/tags/:database", controller.SelectPage)
	api.GET("/tags/:database/:code", controller.SelectPage)
	api.POST("/tags", controller.CreateList)
	api.GET("/tag/:id", controller.SelectById)
	api.POST("/tag", controller.Create)
	api.PUT("/tag", controller.Update)
	api.DELETE("/tag/:id", controller.Delete)

	db := api.Group("/database")
	db.GET("", database.SelectAll)
	db.GET("/:id", database.Select)
	db.POST("", database.Create)
	db.PUT("", database.Update)
	db.DELETE("/:id", database.Delete)

    api.GET("/isonline", influx.ConnectionState)
    api.GET("/status", influx.StatusInfo)

	static := router.Group("/fast")
	static.Static("", "G:\\code\\influxdb-site\\app\\dist")
	return router
}

func initialize(config *models.Config) {
	controller.InitSnapshot(config.Delay)
}
