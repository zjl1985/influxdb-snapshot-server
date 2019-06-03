package main

import (
	"fastdb-server/models"
	"fastdb-server/models/config"
	"fmt"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

var myConfig *models.Config
var engine *xorm.Engine

func main() {
	//if _, err := toml.DecodeFile("config.conf", &myConfig); err != nil {
	//	log.Fatal(err)
	//}
	//r := router.InitRouter(myConfig)
	//myFigure := figure.NewFigure("FastDB", "", true)
	//myFigure.Print()
	//_ = r.Run(myConfig.Port)
	engine, err := xorm.NewEngine("sqlite3", "G:\\sqlite\\data\\rtdb.db")
	if err != nil {
		fmt.Println("--------error--------")
		fmt.Println(err)
	}
	tags := make([]config.Tag, 0)
	databases := make([]config.Database, 0)
	_ = engine.Where("1=1").Find(&databases)
	_ = engine.Find(&tags)
	fmt.Println(tags)
	fmt.Println(databases)
}

func db() *xorm.Engine {
	return engine
}
