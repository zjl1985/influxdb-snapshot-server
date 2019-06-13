package main

import (
    "fastdb-server/router"
    "fastdb-server/service"
    "fmt"
    "github.com/BurntSushi/toml"
    "github.com/common-nighthawk/go-figure"
    "github.com/mattn/go-colorable"
    "github.com/rifflock/lfshook"
    log "github.com/sirupsen/logrus"
)

func main() {
    pathMap := lfshook.PathMap{
        log.DebugLevel: "./log/debug.log",
        log.InfoLevel:  "./log/info.log",
        log.WarnLevel:  "./log/warn.log",
        log.ErrorLevel: "./log/warn.log",
    }
    log.SetFormatter(&log.TextFormatter{ForceColors: true})
    log.SetOutput(colorable.NewColorableStdout())
    log.SetReportCaller(true)
    log.AddHook(lfshook.NewHook(pathMap, &log.TextFormatter{}))
    log.Info("服务启动")
    if _, err := toml.DecodeFile("config.conf", &service.MyConfig); err != nil {
        log.Fatal(err)
    }
    service.MyConfig.FastDBAddress = fmt.Sprintf("http://%s:%s", service.MyConfig.FastDBIP, service.MyConfig.FastDBPort)
    //打开数据库连接
    service.OpenDB()
    if service.MyConfig.Mode == "debug" {
        log.SetLevel(log.DebugLevel)
    } else {
        log.SetLevel(log.InfoLevel)
    }
    //加载路由
    r := router.InitRouter(service.MyConfig)
    //打印欢迎页面
    myFigure := figure.NewFigure("FastDB", "", true)
    myFigure.Print()
    //启动http服务
    _ = r.Run(service.MyConfig.Port)
}
