package actions

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/penalize"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	log "github.com/sirupsen/logrus"
)

func PenalizeNode(node models.Node, repository repositories.NodeRepository) {
	// remove node from active
	err := repository.RemoveNodeFromActive(node)
	if err != nil {
		log.Errorf("Failed penalizing node because of: %v", err)
	}
	// set new cooldown
	node.Cooldown = 60 // initial cooldown is 1 min
	err = repository.Save(&node)
	if err != nil { // todo
		log.Errorf("Failed penalizing node because of: %v", err)
	}
	// schedule new check
	penalize.ScheduleCheckForPenalizedNode(&node, repository)
}