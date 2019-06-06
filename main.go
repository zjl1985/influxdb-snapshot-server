package main

import (
	"fastdb-server/router"
	"fastdb-server/service"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/common-nighthawk/go-figure"
	"github.com/mattn/go-colorable"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})
	log.SetOutput(colorable.NewColorableStdout())
	if _, err := toml.DecodeFile("config.conf", &service.MyConfig); err != nil {
		log.Fatal(err)
	}
	service.MyConfig.FastDBAddress = fmt.Sprintf("http://%s:%s", service.MyConfig.FastDBIP, service.MyConfig.FastDBPort)
	//打开数据库连接
	service.OpenDB()
	if service.MyConfig.Mode == "debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
	//加载路由
	r := router.InitRouter(service.MyConfig)
	//打印欢迎页面
	myFigure := figure.NewFigure("FastDB", "", true)
	myFigure.Print()
	//启动http服务
	_ = r.Run(service.MyConfig.Port)
}
