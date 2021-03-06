package router

import (
    "fastdb-server/controller"
    "fastdb-server/controller/database"
    "fastdb-server/controller/influx"
    "fastdb-server/models"
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    "io"
    "os"
)

// InitRouter 注册服务
func InitRouter(config *models.Config) *gin.Engine {
    gin.SetMode(config.Mode)
    router := gin.New()
    router.Use(gin.Recovery())
    if config.Mode == "debug" {
        router.Use(gin.Logger())
        f, err := os.Create("./log/gin.log")
        if err != nil {
            logrus.Error(err)
        }
        gin.DefaultWriter = io.MultiWriter(f)
    }


    initialize(config)

    router.POST("/snapshot", influx.Snapshot)
    router.POST("/snapshot/write", influx.InfluxSub)
    router.POST("/xinhai/logout", database.LogOut)

    router.POST("/api/login", database.Login)
    api := router.Group("/api")

    if config.EnableAuth {
        api.Use(database.BasicAuth(database.Accounts{
            config.FastUser: config.FastPwd,
        }))
    }
    api.GET("/menu", database.GetMenu)
    api.GET("/getsysinfo", database.GetSysInfo)
    api.GET("/tags", controller.SelectPage)
    api.POST("/tags", controller.CreateList)
    api.GET("/tag/:id", controller.SelectById)
    api.PATCH("/tags-sync/:database", controller.Synchronize)
    api.POST("/tag", controller.Create)
    api.PUT("/tag", controller.Update)
    api.DELETE("/tag/:database/:id", controller.Delete)
    api.DELETE("/tags/:database", controller.DeleteList)

    db := api.Group("/database")
    db.GET("", database.SelectAll)
    db.GET("/:id", database.Select)
    db.POST("", database.Create)
    db.PUT("", database.Update)
    db.DELETE("/:id", database.Delete)

    api.GET("/isonline", influx.ConnectionState)
    api.GET("/status", influx.StatusInfo)

    api.GET("/live", influx.GetLiveData)
    api.GET("/history/data/:database", influx.GeHistoryData)
    api.GET("/history/moment/:database", influx.GeHistoryDataMoment)
    api.GET("/query/:database", influx.UserDefineQuery)

    api.POST("/live/:database", influx.WriteLiveData)
    api.POST("/history/:database", influx.WriteHistoryData)
    api.DELETE("/history/:database", influx.DeleteData)
    api.POST("/upload", influx.ImportData)

    static := router.Group("/fast")
    static.Static("", config.WebPath)
    return router
}

func initialize(config *models.Config) {
    influx.InitSnapshot(config.Delay)
}
