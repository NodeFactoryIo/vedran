package checkactive

import (
	"time"

	"github.com/NodeFactoryIo/vedran/internal/actions"
	"github.com/NodeFactoryIo/vedran/internal/active"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultScheduleInterval = 10 * time.Second
)

// Start scheduled task on DefaultScheduleInterval that checks for each active node if it is active
// and penalizes node if it is not active
func StartScheduledTask(repos *repositories.Repos) {
	ticker := time.NewTicker(DefaultScheduleInterval)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				scheduledTask(repos, actions.NewActions())
			}
		}
	}()
}

func scheduledTask(repos *repositories.Repos, actions actions.Actions) {
	log.Debug("Started task: check all active nodes")
	activeNodes := repos.NodeRepo.GetAllActiveNodes()

	for _, node := range *activeNodes {

		pingActive, err := active.CheckIfPingActive(node.ID, repos)
		if err != nil {
			log.Errorf("Unable to check if node %s active because of %v", node.ID, err)
			continue
		}

		if !pingActive {
			actions.PenalizeNode(node, *repos)
			continue
		}

		metricsValid, err := active.CheckIfMetricsValid(node.ID, repos)
		if err != nil {
			log.Errorf("Unable to check if node %s active because of %v", node.ID, err)
			continue
		}

		if !metricsValid {
			err = repos.NodeRepo.RemoveNodeFromActive(node.ID)
			if err != nil {
				log.Errorf("Unable to remove node %s from active because of %v", node.ID, err)
			}
			log.Debugf("Node %s metrics lagging more than 10 blocks, removed node from active", node.ID)
		}
	}
}
