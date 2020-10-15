package actions

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/penalize"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	log "github.com/sirupsen/logrus"
)

const InitialPenalizeIntervalInSeconds = 60

func PenalizeNode(node models.Node, repository repositories.NodeRepository) {
	// remove node from active
	err := repository.RemoveNodeFromActive(node)
	if err != nil {
		log.Errorf("Failed penalizing node because of: %v", err)
		return
	}

	// set new cooldown
	node.Cooldown = InitialPenalizeIntervalInSeconds // initial cooldown is 1 min
	err = repository.Save(&node)
	if err != nil {
		log.Errorf("Failed penalizing node because of: %v", err)
	}

	log.Debugf("Penalized node %s, on cooldown for 1 minute ", node.ID)
	penalize.ScheduleCheckForPenalizedNode(&node, repository)
}