package actions

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"github.com/NodeFactoryIo/vedran/internal/schedule/penalize"
	log "github.com/sirupsen/logrus"
)

const InitialPenalizeIntervalInMins = 1

// PenalizeNode removes provided node from active nodes, sets initial cooldown of 1 minute and schedules check for
// penalized node by invoking penalize.ScheduleCheckForPenalizedNode
func (a *actions) PenalizeNode(node models.Node, repositories repositories.Repos, message string) {
	// remove node from active
	err := repositories.NodeRepo.RemoveNodeFromActive(node.ID)
	if err != nil {
		log.Errorf("Failed penalizing node %s because of: %v", node.ID, err)
		return
	}

	// set new cooldown
	node.Cooldown = InitialPenalizeIntervalInMins
	err = repositories.NodeRepo.Save(&node)
	if err != nil {
		log.Errorf("Failed penalizing node %s because of: %v", node.ID, err)
		return
	}

	log.Debugf("Penalized node %s, on cooldown for 1 minute, because %s ", node.ID, message)
	go penalize.ScheduleCheckForPenalizedNode(node, repositories)
}
