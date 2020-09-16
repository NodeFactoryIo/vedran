package db

import "github.com/asdine/storm/v3"

type Node struct {
	ID            int    `storm:"id"`
	ConfigHash    string `storm:"config_hash"`
	NodeUrl       string `storm:"config_hash"`
	PayoutAddress string `storm:"payout_address"`
	Token         string `storm:"token"`
}

type NodeDatabaseService interface {

}

type nodeDatabaseService struct {
	db *storm.DB
}

func (nds *nodeDatabaseService) SaveNode()  {

}
