package repositories

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/asdine/storm/v3"
)

type NodeRepo struct {
	db *storm.DB
}

func NewNodeRepo(db *storm.DB) *NodeRepo {
	return &NodeRepo{
		db: db,
	}
}

func (r *NodeRepo) FindByID(ID string) (*models.Node, error) {
	var node *models.Node
	err := r.db.One("ID", ID, node)
	return node, err
}

func (r *NodeRepo) Save(node *models.Node) error {
	return r.db.Save(node)
}

func (r *NodeRepo) GetAll() (*[]models.Node, error) {
	var nodes []models.Node
	err := r.db.All(&nodes)
	return &nodes, err
}

func (r *NodeRepo) IsNodeWhitelisted(ID string) (bool, error) {
	var isWhitelisted bool
	err := r.db.Get(models.WhitelistBucket, ID, &isWhitelisted)
	return isWhitelisted, err
}