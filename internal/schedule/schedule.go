package schedule

import (
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/actions"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	ScheduleInterval = 5 * time.Second
	IntervalFromLastPing = 10 * time.Second
)

func StartScheduleTask(repos *repositories.Repos) {
	ticker := time.NewTicker(ScheduleInterval)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				scheduledTask(repos)
			}
		}
	}()
}

func scheduledTask(repos *repositories.Repos) {
	log.Debugf("%v SCHEDULED TASK\n", time.Now())
	active := repos.NodeRepo.GetAllActiveNodes()

	fmt.Println("ACTIVE NODES")
	for _, node := range *active {
		lastPing, err := repos.PingRepo.FindByNodeID(node.ID)
		if err != nil {
			log.Error(err)
		}

		fmt.Printf("%s NODE: %s\n", node.ID, lastPing.Timestamp.String())

		// more than 10 seconds passed from last ping
		if lastPing.Timestamp.Add(IntervalFromLastPing).Before(time.Now()) {
			log.Infof("PENALIZE NODE %s", node.ID)
			actions.PenalizeNode(node, repos.NodeRepo)
			return
		}

		// node's latest and best block lag behind the best in the pool by more than 10 blocks
		metrics, err := repos.MetricsRepo.FindByID(node.ID)
		if err != nil {
			return // todo
		}
		latestBlockMetrics, err := repos.MetricsRepo.GetLatestBlockMetrics()
		if err != nil {
			return // todo
		}
		if metrics.BestBlockHeight <= latestBlockMetrics.BestBlockHeight-10 &&
			metrics.FinalizedBlockHeight <= latestBlockMetrics.FinalizedBlockHeight-10 {
			log.Infof("PENALIZE NODE %s", node.ID)
			actions.PenalizeNode(node, repos.NodeRepo)
		}
	}
}