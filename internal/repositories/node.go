package repositories

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
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

func (r *NodeRepo) getValidNodes() (*[]models.Node, error) {
	var nodes []models.Node

	q := r.db.Select(q.Lte("Cooldown", 0))
	err := q.Find(&nodes)

	return &nodes, err
}

func (r *NodeRepo) InitNodeRepo() error {
	nodes, err := r.getValidNodes()

	if err != nil {
		if err.Error() == "not found" {
			memoryNodes = make([]models.Node, 0)
			return nil
		}

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
	r.AddNodeToActive(*node)

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
	nodes := make([]models.Node, len(memoryNodes))

	_ = copy(nodes[:], memoryNodes)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(nodes), func(i, j int) {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})

	return &nodes
}

func (r *NodeRepo) getRoundRobinNodes() *[]models.Node {
	nodes := make([]models.Node, len(memoryNodes))

	_ = copy(nodes[:], memoryNodes)

	sort.Slice(nodes[:], func(i, j int) bool {
		return nodes[i].LastUsed < nodes[j].LastUsed
	})

	return &nodes
}

func (r *NodeRepo) GetActiveNodes(selection string) *[]models.Node {
	if selection == "round-robin" {
		return r.getRoundRobinNodes()
	}

	return r.getRandomNodes()
}

func (r *NodeRepo) updateMemoryLastUsedTime(targetNode models.Node) {
	for i, node := range memoryNodes {
		if targetNode.ID == node.ID {
			tempNode := &memoryNodes[i]
			tempNode.LastUsed = time.Now().Unix()
			break
		}
	}
}

func (r *NodeRepo) RemoveNodeFromActive(targetNode models.Node) error {
	for i, node := range memoryNodes {

		if targetNode.ID == node.ID {
			memoryNodes[i] = memoryNodes[len(memoryNodes)-1]
			memoryNodes = memoryNodes[:len(memoryNodes)-1]
			return nil
		}

	}

	return fmt.Errorf("No target node %s in memory", targetNode.ID)
}

func (r *NodeRepo) AddNodeToActive(node models.Node) {
	memoryNodes = append(memoryNodes, node)
}

func (r *NodeRepo) PenalizeNode(node models.Node) {
	err := r.RemoveNodeFromActive(node)
	if err != nil {
		log.Errorf("Failed penalizing node because of: %v", err)
	}
}

func (r *NodeRepo) RewardNode(node models.Node) {
	r.updateMemoryLastUsedTime(node)

	node.LastUsed = time.Now().Unix()
	err := r.db.Update(&node)
	if err != nil {
		log.Errorf("Failed updating node last used time because of: %v", err)
	}
}
