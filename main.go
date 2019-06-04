package main

import (
	"fastdb-server/models"
	"fastdb-server/router"
	"fastdb-server/service"
	"github.com/BurntSushi/toml"
	"github.com/common-nighthawk/go-figure"
	"github.com/mattn/go-colorable"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

var myConfig *models.Config

func main() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})
	log.SetOutput(colorable.NewColorableStdout())
	if _, err := toml.DecodeFile("config.conf", &myConfig); err != nil {
		log.Fatal(err)
	}
	//打开数据库连接
	service.OpenDB(myConfig)
	if myConfig.Mode == "debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
	//加载路由
	r := router.InitRouter(myConfig)
	//打印欢迎页面
	myFigure := figure.NewFigure("FastDB", "", true)
	myFigure.Print()
	//启动http服务
	_ = r.Run(myConfig.Port)
}
