package penalize

import (
	"github.com/NodeFactoryIo/vedran/internal/active"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"github.com/NodeFactoryIo/vedran/internal/whitelist"
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
			return
		}

		if isActive {
			nodeWithResetedCooldown, err := repositories.NodeRepo.ResetNodeCooldown(node.ID)
			if err != nil {
				log.Errorf("Unable to reset node %s cooldown, because of %v", node.ID, err)
				return
			}
			err = repositories.NodeRepo.AddNodeToActive(*nodeWithResetedCooldown)
			if err != nil {
				log.Errorf("Unable to set node %s as active, because of %v", node.ID, err)
			}
		} else {
			nodeWithNewCooldown, err := repositories.NodeRepo.IncreaseNodeCooldown(node.ID)
			if err != nil {
				log.Errorf("Unable to save new cooldown for node %s, because of %v", node.ID, err)
				return
			}
			if (time.Duration(nodeWithNewCooldown.Cooldown) * time.Minute) > MaxCooldownForPenalizedNode {
				log.Debugf("Node %s reached maximum cooldown", node.ID)
				err = whitelist.RemoveNodeFromWhitelisted(node.ID)
				if err != nil {
					log.Errorf("Unable to remove node %s from whitelisted nodes, because of %v", node.ID, err)
				}
				return
			}
			ScheduleCheckForPenalizedNode(*nodeWithNewCooldown, repositories)
		}
	})
}