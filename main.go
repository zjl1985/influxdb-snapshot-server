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
    "strconv"
    "sync"
    "time"
)

func main() {

    start()
    //test()
}

func test() {
    //var list chan []string
    //list = make(chan []string, 10)
    var wg sync.WaitGroup
    result := make([]string, 0)
    queue := make(chan string, 1)
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go write(strconv.Itoa(i), queue)
        //restul := <-list
        //logrus.Info(restul)
    }
    go func() {
        // defer wg.Done() <- Never gets called since the 100 `Done()` calls are made above, resulting in the `Wait()` to continue on before this is executed
        for t := range queue {
            result = append(result, t)
            wg.Done() // ** move the `Done()` call here
        }
    }()

    wg.Wait()
    logrus.Info(len(result))

}

func write(index string, output chan string) {
    time.Sleep(time.Second)
    output <- index
}

func start() {
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
    if _, err := toml.DecodeFile("fastdb-snapshot.conf", &service.MyConfig); err != nil {
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
