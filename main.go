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
	service.OpenDB(myConfig)

	r := router.InitRouter(myConfig)
	myFigure := figure.NewFigure("FastDB", "", true)
	myFigure.Print()
	_ = r.Run(myConfig.Port)
}
