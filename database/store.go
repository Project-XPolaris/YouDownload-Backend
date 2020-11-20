package database

import (
	"github.com/asdine/storm"
	log "github.com/sirupsen/logrus"
)
var Instance *storm.DB
func InitDB(dbPath string) {
	var err error
	Instance, err = storm.Open(dbPath)
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "Path": dbPath}).Fatal("Failed to create database")
	}
}