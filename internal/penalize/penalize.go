package penalize

import (
	"github.com/NodeFactoryIo/vedran/internal/active"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"time"
)

const MaxCooldown = 1000000

func ScheduleCheckForPenalizedNode(node models.Node, repositories repositories.Repos) {
	time.AfterFunc(time.Duration(node.Cooldown), func() {
		isActive, err := active.CheckIfNodeActive(node, &repositories)
		if err != nil {

		}

		if isActive {
			repositories.NodeRepo.AddNodeToActive(node)
		} else {
			newCooldown := node.Cooldown * 2
			if newCooldown > MaxCooldown {
				// TODO
			}
			node.Cooldown = newCooldown
			err := repositories.NodeRepo.Save(&node)
			if err != nil {
				// TODO
			}
			ScheduleCheckForPenalizedNode(node, repositories)
		}
	})
}