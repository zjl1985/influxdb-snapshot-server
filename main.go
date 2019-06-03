package main

import (
	"fastdb-server/models"
	"github.com/BurntSushi/toml"
	"github.com/common-nighthawk/go-figure"
	"log"
)

var config *models.Config

func main() {
	if _, err := toml.DecodeFile("config.conf", &config); err != nil {
		log.Fatal(err)
	}
	router := initRouter(config)
	myFigure := figure.NewFigure("FastDB", "", true)
	myFigure.Print()
	_ = router.Run(config.Port)
}
