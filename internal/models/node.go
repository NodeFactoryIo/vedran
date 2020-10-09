package models

type Node struct {
	ID            string `storm:"id"`
	ConfigHash    string
	NodeUrl       string
	PayoutAddress string
	Token         string
	Cooldown      int
	LastUsed      int64
}

type NodeRepository interface {
	FindByID(ID string) (*Node, error)
	Save(node *Node) error
	GetAll() (*[]Node, error)
	GetActiveNodes(selection string) *[]Node
	RemoveNodeFromActive(node *Node) error
	AddNodeToActive(node *Node)
	IsNodeWhitelisted(ID string) (bool, error)
	PenalizeNode(node *Node)
	RewardNode(node *Node)
}
