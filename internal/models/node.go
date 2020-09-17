package models

type Node struct {
	ID            string `storm:"id"`
	ConfigHash    string
	NodeUrl       string
	PayoutAddress string
	Token         string
}

type NodeRepository interface {
	FindByID(ID int) (*Node, error)
	Save(node *Node) error
	GetAll() (*[]Node, error)
}
