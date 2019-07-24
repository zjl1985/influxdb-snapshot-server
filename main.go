package main

import (
    "fastdb-server/controller/influx"
    "fastdb-server/router"
    "fastdb-server/service"
    "fmt"
    "github.com/BurntSushi/toml"
    "github.com/common-nighthawk/go-figure"
    "github.com/mattn/go-colorable"
    "github.com/rifflock/lfshook"
    "github.com/sirupsen/logrus"
    "log"
)

func main() {
    pathMap := lfshook.PathMap{
        logrus.DebugLevel: "./log/debug.log",
        logrus.InfoLevel:  "./log/info.log",
        logrus.WarnLevel:  "./log/warn.log",
        logrus.ErrorLevel: "./log/warn.log",
    }
    logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
    logrus.SetOutput(colorable.NewColorableStdout())
    logrus.SetReportCaller(true)
    logrus.AddHook(lfshook.NewHook(pathMap, &logrus.TextFormatter{}))
    logrus.Info("服务启动")
    if _, err := toml.DecodeFile("config.conf", &service.MyConfig); err != nil {
        log.Fatal(err)
    }
    service.MyConfig.FastDBAddress = fmt.Sprintf("http://%s:%s", service.MyConfig.FastDBIP, service.MyConfig.FastDBPort)
    influx.CreatAdmin()

    //打开数据库连接
    service.OpenDB()
    if service.MyConfig.Mode == "debug" {
        logrus.SetLevel(logrus.DebugLevel)
    } else {
        logrus.SetLevel(logrus.InfoLevel)
    }
    //加载路由
    r := router.InitRouter(service.MyConfig)
    //打印欢迎页面
    myFigure := figure.NewFigure("FastDB", "", true)
    myFigure.Print()
    //启动http服务
    _ = r.Run(service.MyConfig.Port)
}
