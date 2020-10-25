package repositories

import (
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/asdine/storm/v3"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"sort"
	"time"
)

var activeNodes []models.Node

type NodeRepository interface {
	FindByID(ID string) (*models.Node, error)
	Save(node *models.Node) error
	GetAll() (*[]models.Node, error)
	GetActiveNodes(selection string) *[]models.Node
	GetAllActiveNodes() *[]models.Node
	RemoveNodeFromActive(ID string) error
	AddNodeToActive(ID string) error
	RewardNode(node models.Node)
	IncreaseNodeCooldown(ID string) (*models.Node, error)
	ResetNodeCooldown(ID string) (*models.Node, error)
	IsNodeOnCooldown(ID string) (bool, error)
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

func (r *nodeRepo) RemoveNodeFromActive(ID string) error {
	for i, node := range activeNodes {
		if ID == node.ID {
			activeNodes[i] = activeNodes[len(activeNodes)-1]
			activeNodes = activeNodes[:len(activeNodes)-1]
			return nil
		}
	}

	return fmt.Errorf("no target node %s in memory", ID)
}

func (r *nodeRepo) AddNodeToActive(ID string) error {
	// check if node exists
	node, err := r.FindByID(ID)
	if err != nil {
		return err
	}
	// check if already active
	for _, activeNode := range activeNodes {
		if activeNode.ID == ID {
			return fmt.Errorf("node %s already set as active", ID)
		}
	}

	activeNodes = append(activeNodes, *node)
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

// IsNodeOnCooldown check if node is on cooldown
func (r *nodeRepo) IsNodeOnCooldown(ID string) (bool, error) {
	var node models.Node
	err := r.db.One("ID", ID, &node)
	if err != nil {
		return false, err
	}

	return node.Cooldown != 0, err
}

