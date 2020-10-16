package penalize

import (
	"github.com/NodeFactoryIo/vedran/internal/active"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	log "github.com/sirupsen/logrus"
	"time"
)

const MaxCooldownForPenalizedNode = 17 * time.Hour

var afterFunc = time.AfterFunc

func ScheduleCheckForPenalizedNode(node models.Node, repositories repositories.Repos) {
	afterFunc(time.Duration(node.Cooldown), func() {
		isActive, err := active.CheckIfNodeActive(node, &repositories)
		if err != nil {
			log.Errorf("Unable to check if node %s active, because of %v", node.ID, err)
		}

		if isActive {
			repositories.NodeRepo.AddNodeToActive(node)
		} else {
			newCooldown := node.Cooldown * 2
			if (time.Duration(newCooldown) * time.Minute) > MaxCooldownForPenalizedNode {
				log.Debugf("Node %s reached maximum cooldown", node.ID)
				// TODO - remove node
				return
			}
			node.Cooldown = newCooldown
			err := repositories.NodeRepo.Save(&node)
			if err != nil {
				log.Errorf("Unable to save new cooldown for node %s, because of %v", node.ID, err)
			}
			ScheduleCheckForPenalizedNode(node, repositories)
		}
	})
}