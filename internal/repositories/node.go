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

var activeNodes []models.Node

type NodeRepository interface {
	InitNodeRepo() error
	FindByID(ID string) (*models.Node, error)
	Save(node *models.Node) error
	GetAll() (*[]models.Node, error)
	IsNodeWhitelisted(ID string) (bool, error)
	GetActiveNodes(selection string) *[]models.Node
	GetAllActiveNodes() *[]models.Node
	RemoveNodeFromActive(node models.Node) error
	AddNodeToActive(node models.Node)
	RewardNode(node models.Node)
}

type nodeRepo struct {
	db *storm.DB
}

func NewNodeRepo(db *storm.DB) NodeRepository {
	return &nodeRepo{
		db: db,
	}
}

func (r *nodeRepo) getValidNodes() (*[]models.Node, error) {
	var nodes []models.Node

	q := r.db.Select(q.Lte("Cooldown", 0))
	err := q.Find(&nodes)

	return &nodes, err
}

func (r *nodeRepo) InitNodeRepo() error {
	nodes, err := r.getValidNodes()
	if err != nil {
		return err
	}

	activeNodes = *nodes
	return nil
}

func (r *nodeRepo) FindByID(ID string) (*models.Node, error) {
	var node *models.Node
	err := r.db.One("ID", ID, node)
	return node, err
}

func (r *nodeRepo) Save(node *models.Node) error {
	r.AddNodeToActive(*node)

	return r.db.Save(node)
}

func (r *nodeRepo) GetAll() (*[]models.Node, error) {
	var nodes []models.Node
	err := r.db.All(&nodes)
	return &nodes, err
}

func (r *nodeRepo) IsNodeWhitelisted(ID string) (bool, error) {
	var isWhitelisted bool
	err := r.db.Get(models.WhitelistBucket, ID, &isWhitelisted)
	return isWhitelisted, err
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

func (r *nodeRepo) AddNodeToActive(node models.Node) {
	activeNodes = append(activeNodes, node)
}

func (r *nodeRepo) RewardNode(node models.Node) {
	r.updateMemoryLastUsedTime(node)

	node.LastUsed = time.Now().Unix()
	err := r.db.Update(node)
	if err != nil {
		log.Errorf("Failed updating node last used time because of: %v", err)
	}
}
