package db

import (
	"github.com/asdine/storm/v3"
	"log"
)

type DatabaseService struct {
	DB *storm.DB
}

var databaseServiceInstance *DatabaseService

func newDatabaseService() *DatabaseService {
	db, err := storm.Open("my.db")
	if err != nil {
		// TODO handle error
		log.Print(err)
	}

	defer db.Close()
	return &DatabaseService{
		DB: db,
	}

}

func GetDatabaseService() *DatabaseService {
	if databaseServiceInstance == nil {
		databaseServiceInstance = newDatabaseService()
	}
	return databaseServiceInstance
}

