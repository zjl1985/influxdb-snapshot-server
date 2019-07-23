package service

import (
    "fastdb-server/models"
    "github.com/go-xorm/core"
    "github.com/go-xorm/xorm"
    _ "github.com/mattn/go-sqlite3"
    "github.com/sirupsen/logrus"
    "os"
)

var Engine *xorm.Engine
var MyConfig *models.Config

func OpenDB() {
    f, err := os.Create("./log/sql.log")
    if err != nil {
        logrus.Error(err)
    }
    Engine, _ = xorm.NewEngine("sqlite3", MyConfig.DBPath)
    cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
    Engine.SetDefaultCacher(cacher)
    Engine.Logger().SetLevel(core.LOG_WARNING)

    if MyConfig.Mode == "debug" {
        Engine.ShowSQL(true)
        Engine.Logger().SetLevel(core.LOG_DEBUG)
    }
    Engine.SetLogger(xorm.NewSimpleLogger(f))
    err = Engine.Ping()
    if err != nil {
        logrus.Fatal(err)
    }
}
