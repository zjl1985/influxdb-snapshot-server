package service

import (
	"fastdb-server/models"
	"github.com/go-xorm/xorm"
	log "github.com/sirupsen/logrus"
)

var Engine *xorm.Engine

func OpenDB(config *models.Config) {
	Engine, _ = xorm.NewEngine("sqlite3", config.DBPath)
	if config.Mode == "debug" {
		Engine.ShowSQL(true)
	}
	err := Engine.Ping()
	if err != nil {
		log.Fatal(err)
	}
}
