package repositories

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/asdine/storm/v3"
	log "github.com/sirupsen/logrus"
)

var memoryNodes []models.Node

type NodeRepo struct {
	db *storm.DB
}

func NewNodeRepo(db *storm.DB) *NodeRepo {
	return &NodeRepo{
		db: db,
	}
}

func (r *NodeRepo) InitNodeRepo() error {
	nodes, err := r.GetAll()
	if err != nil {
		return err
	}

	memoryNodes = *nodes
	return nil
}

func (r *NodeRepo) FindByID(ID string) (*models.Node, error) {
	var node *models.Node
	err := r.db.One("ID", ID, node)
	return node, err
}

func (r *NodeRepo) Save(node *models.Node) error {
	r.AddNodeToActive(node)

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

func (r *NodeRepo) getRandomNodes() *[]models.Node {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(memoryNodes), func(i, j int) {
		memoryNodes[i], memoryNodes[j] = memoryNodes[j], memoryNodes[i]
	})

	return &memoryNodes
}

func (r *NodeRepo) getRoundRobinNodes() *[]models.Node {
	sort.Slice(memoryNodes[:], func(i, j int) bool {
		return memoryNodes[i].LastUsed < memoryNodes[j].LastUsed
	})

	return &memoryNodes
}

func (r *NodeRepo) GetActiveNodes(selection string) *[]models.Node {
	if selection == "round-robin" {
		return r.getRoundRobinNodes()
	}

	return r.getRandomNodes()
}

func (r *NodeRepo) updateMemoryLastUsedTime(targetNode *models.Node) error {
	for i, node := range memoryNodes {
		if targetNode.ID == node.ID {
			tempNode := &memoryNodes[i]
			tempNode.LastUsed = time.Now().Unix()
			return nil
		}
	}

	return fmt.Errorf("No target node in memory")
}

func (r *NodeRepo) RemoveNodeFromActive(targetNode *models.Node) error {
	for i, node := range memoryNodes {

		if targetNode.ID == node.ID {
			memoryNodes[len(memoryNodes)-1], memoryNodes[i] = memoryNodes[i], memoryNodes[len(memoryNodes)-1]
			memoryNodes = memoryNodes[:len(memoryNodes)-1]
			return nil
		}

	}

	return fmt.Errorf("No target node in memory")
}

func (r *NodeRepo) AddNodeToActive(node *models.Node) {
	memoryNodes = append(memoryNodes, *node)
}

func (r *NodeRepo) PenalizeNode(node *models.Node) {
	_ = r.RemoveNodeFromActive(node)
}

func (r *NodeRepo) RewardNode(node *models.Node) {
	_ = r.updateMemoryLastUsedTime(node)

	node.LastUsed = time.Now().Unix()
	err := r.db.Update(node)
	if err != nil {
		log.Errorf("Failed penalizing node because of: %v", err)
	}
}
