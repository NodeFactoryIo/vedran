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

var activeNodes []models.Node

type NodeRepository interface {
	FindByID(ID string) (*models.Node, error)
	Save(node *models.Node) error
	GetAll() (*[]models.Node, error)
	GetActiveNodes(selection string) *[]models.Node
	GetAllActiveNodes() *[]models.Node
	RemoveNodeFromActive(node models.Node) error
	AddNodeToActive(node models.Node) error
	RewardNode(node models.Node)
	IncreaseNodeCooldown(ID string) (*models.Node, error)
	ResetNodeCooldown(ID string) (*models.Node, error)
}

type nodeRepo struct {
	db *storm.DB
}

func NewNodeRepo(db *storm.DB) NodeRepository {
	activeNodes = make([]models.Node, 0)

	return &nodeRepo{
		db: db,
	}
}

func (r *nodeRepo) FindByID(ID string) (*models.Node, error) {
	var node models.Node
	err := r.db.One("ID", ID, &node)
	return &node, err
}

func (r *nodeRepo) Save(node *models.Node) error {
	err := r.db.Save(node)
	if err != nil {
		return err
	}
	return nil
}

func (r *nodeRepo) GetAll() (*[]models.Node, error) {
	var nodes []models.Node
	err := r.db.All(&nodes)
	return &nodes, err
}

func (r *nodeRepo) getRandomNodes() *[]models.Node {
	nodes := make([]models.Node, len(activeNodes))

	_ = copy(nodes[:], activeNodes)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(nodes), func(i, j int) {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})

	return &nodes
}

func (r *nodeRepo) getRoundRobinNodes() *[]models.Node {
	nodes := make([]models.Node, len(activeNodes))

	_ = copy(nodes[:], activeNodes)

	sort.Slice(nodes[:], func(i, j int) bool {
		return nodes[i].LastUsed < nodes[j].LastUsed
	})

	return &nodes
}

func (r *nodeRepo) GetActiveNodes(selection string) *[]models.Node {
	if selection == "round-robin" {
		return r.getRoundRobinNodes()
	}

	return r.getRandomNodes()
}

func (r *nodeRepo) GetAllActiveNodes() *[]models.Node {
	return &activeNodes
}

func (r *nodeRepo) updateMemoryLastUsedTime(targetNode models.Node) {
	for i, node := range activeNodes {
		if targetNode.ID == node.ID {
			tempNode := &activeNodes[i]
			tempNode.LastUsed = time.Now().Unix()
			break
		}
	}
}

func (r *nodeRepo) RemoveNodeFromActive(targetNode models.Node) error {
	for i, node := range activeNodes {
		if targetNode.ID == node.ID {
			activeNodes[i] = activeNodes[len(activeNodes)-1]
			activeNodes = activeNodes[:len(activeNodes)-1]
			return nil
		}
	}

	return fmt.Errorf("No target node %s in memory", targetNode.ID)
}

func (r *nodeRepo) AddNodeToActive(node models.Node) error {
	for _, activeNode := range activeNodes {
		if activeNode.ID == node.ID {
			return fmt.Errorf("node %s already set as active", node.ID)
		}
	}
	activeNodes = append(activeNodes, node)
	return nil
}

func (r *nodeRepo) RewardNode(node models.Node) {
	r.updateMemoryLastUsedTime(node)

	node.LastUsed = time.Now().Unix()
	err := r.db.Update(&node)
	if err != nil {
		log.Errorf("Failed updating node last used time because of: %v", err)
	}
}

// IncreaseNodeCooldown doubles node cooldown and saves it to db
func (r *nodeRepo) IncreaseNodeCooldown(ID string) (*models.Node, error) {
	var node models.Node
	err := r.db.One("ID", ID, &node)
	if err != nil {
		return nil, err
	}

	newCooldown := 2 * node.Cooldown
	node.Cooldown = newCooldown

	err = r.db.Save(&node)
	return &node, err
}

// ResetNodeCooldown resets node cooldown to 0 and saves it to db
func (r *nodeRepo) ResetNodeCooldown(ID string) (*models.Node, error) {
	var node models.Node
	err := r.db.One("ID", ID, &node)
	if err != nil {
		return nil, err
	}

	node.Cooldown = 0

	err = r.db.Save(&node)
	return &node, err
}
