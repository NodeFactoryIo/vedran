package checkactive

import (
	"github.com/NodeFactoryIo/vedran/internal/actions"
	"github.com/NodeFactoryIo/vedran/internal/active"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	DefaultScheduleInterval = 10 * time.Second
)

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
		isActive, err := active.CheckIfNodeActive(node, repos)
		if err != nil {
			log.Errorf("Unable to check if node %s active because of %v", node.ID, err)
		}
		if !isActive {
			actions.PenalizeNode(node, *repos)
		}
	}
}