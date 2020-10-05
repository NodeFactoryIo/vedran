package models

type Node struct {
	ID            string `storm:"id"`
	ConfigHash    string
	NodeUrl       string
	PayoutAddress string
	Token         string
	Cooldown      int
}

type NodeRepository interface {
	FindByID(ID string) (*Node, error)
	Save(node *Node) error
	GetAll() (*[]Node, error)
	GetActiveNodes() (*[]Node, error)
	IsNodeWhitelisted(ID string) (bool, error)
}
