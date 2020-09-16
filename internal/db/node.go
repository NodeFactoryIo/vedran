package db

import "github.com/asdine/storm/v3"

type Node struct {
	ID            int    `storm:"id"`
	ConfigHash    string
	NodeUrl       string
	PayoutAddress string `storm:"index"`
	Token         string
}

type NodeDatabaseService interface {

}

type nodeDatabaseService struct {
	db *storm.DB
}

func (nds *nodeDatabaseService) SaveNode()  {

}
