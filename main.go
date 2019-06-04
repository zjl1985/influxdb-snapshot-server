package main

import (
	"fastdb-server/models"
	"fastdb-server/router"
	"fastdb-server/service"
	"github.com/BurntSushi/toml"
	"github.com/common-nighthawk/go-figure"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var myConfig *models.Config

func main() {
	if _, err := toml.DecodeFile("config.conf", &myConfig); err != nil {
		log.Fatal(err)
	}
	//打开数据库连接
	service.OpenDB(myConfig)
	//加载路由
	r := router.InitRouter(myConfig)
	//打印欢迎页面
	myFigure := figure.NewFigure("FastDB", "", true)
	myFigure.Print()
	//启动http服务
	_ = r.Run(myConfig.Port)
}
