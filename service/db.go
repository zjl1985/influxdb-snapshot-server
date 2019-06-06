package service

import (
	"fastdb-server/models"
	"github.com/go-xorm/xorm"
	log "github.com/sirupsen/logrus"
)

var Engine *xorm.Engine
var MyConfig *models.Config

func OpenDB() {
	Engine, _ = xorm.NewEngine("sqlite3", MyConfig.DBPath)
	if MyConfig.Mode == "debug" {
		Engine.ShowSQL(true)
	}
	err := Engine.Ping()
	if err != nil {
		log.Fatal(err)
	}
}
