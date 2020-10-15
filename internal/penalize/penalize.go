package penalize

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	log "github.com/sirupsen/logrus"
	"time"
)

func ScheduleCheckForPenalizedNode(node *models.Node, nodeRepo repositories.NodeRepository) {
	time.AfterFunc(time.Duration(node.Cooldown), func() {
		// TODO check if node active

		// YES ->
		// nodeRepo.AddNodeToActive(*node)

		// NO ->
		// newCooldown := node.Cooldown * 2
		// TODO check if cooldown bigger than MAX_COOLDOWN
		// node.Cooldown = newCooldown
		// _ = nodeRepo.Save(node) // todo error
		// schedule new check
		// ScheduleCheckForPenalizedNode(node, nodeRepo)

		log.Infof("CHECK FOR PENALIZED NODE %s", node.ID)
	})
}