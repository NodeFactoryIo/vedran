package models

import "time"

type Node struct {
	ID            string `storm:"id"`
	ConfigHash    string
	NodeUrl       string
	PayoutAddress string
	Token         string
	Cooldown      int
	LastUsed      time.Time
}

type NodeRepository interface {
	FindByID(ID string) (*Node, error)
	Save(node *Node) error
	GetAll() (*[]Node, error)
	GetActiveNodes(selection string) (*[]Node, error)
	IsNodeWhitelisted(ID string) (bool, error)
	PenalizeNode(node *Node) error
	RewardNode(node *Node) error
}
