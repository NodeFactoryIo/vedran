package active

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	IntervalFromLastPing = 10 * time.Second
	AllowedBlocksBehind  = 10
	TargetBlockBuffer    = 10
)

// CheckIfNodeActive checks if nodes last recorded ping is in last IntervalFromLastPing and if nodes last recorded
// BestBlockHeight and FinalizedBlockHeight are lagging more than AllowedBlocksBehind blocks
func CheckIfNodeActive(node models.Node, repos *repositories.Repos) (bool, error) {
	isPingActive, err := CheckIfPingActive(node.ID, repos)
	if !isPingActive {
		return false, err
	}

	isMetricsValid, err := CheckIfMetricsValid(node.ID, repos)
	if !isMetricsValid {
		return false, err
	}

	return true, nil
}

// CheckIfPingActive checks if nodes last recorded ping is in last IntervalFromLastPing
func CheckIfPingActive(nodeID string, repos *repositories.Repos) (bool, error) {
	lastPing, err := repos.PingRepo.FindByNodeID(nodeID)
	if err != nil {
		return false, err
	}

	// more than 10 seconds passed from last ping
	if lastPing.Timestamp.Add(IntervalFromLastPing).Before(time.Now()) {
		log.Debugf("Node %s not active as last ping was at %v", nodeID, lastPing.Timestamp)
		return false, nil
	}

	return true, nil
}

// CheckIfMetricsValid checks if nodes last recorded BestBlockHeight and FinalizedBlockHeight
// are lagging more than AllowedBlocksBehind blocks
func CheckIfMetricsValid(nodeID string, repos *repositories.Repos) (bool, error) {
	metrics, err := repos.MetricsRepo.FindByID(nodeID)
	if err != nil {
		return false, err
	}
	// check if node synced
	if (metrics.BestBlockHeight + TargetBlockBuffer) < metrics.TargetBlockHeight {
		log.Debugf(
			"Node %s not synced. Best block: %d, Target block: %d",
			nodeID, metrics.BestBlockHeight, metrics.TargetBlockHeight,
		)
		return false, nil
	}
	// get best metrics from pool of nodes
	latestBlockMetrics, err := repos.MetricsRepo.GetLatestBlockMetrics()
	if err != nil {
		return false, err
	}
	// check if node falling behind
	if metrics.BestBlockHeight <= (latestBlockMetrics.BestBlockHeight-AllowedBlocksBehind) ||
		metrics.FinalizedBlockHeight <= (latestBlockMetrics.FinalizedBlockHeight-AllowedBlocksBehind) {
		log.Debugf(
			"Node %s not active as metrics check failed. "+
				"Node metrics: BestBlockHeight[%d], FinalizedBlockHeight[%d] "+
				"Best pool metrics: BestBlockHeight[%d], FinalizedBlockHeight[%d]",
			nodeID,
			metrics.BestBlockHeight, metrics.FinalizedBlockHeight,
			latestBlockMetrics.BestBlockHeight, latestBlockMetrics.FinalizedBlockHeight,
		)
		return false, nil
	}
	return true, nil
}

// ActivateNodeIfReady adds node to active nodes if latest metrics are valid and node is not penalized
func ActivateNodeIfReady(nodeID string, repos repositories.Repos) error {
	nodeIsOnCooldown, err := repos.NodeRepo.IsNodeOnCooldown(nodeID)
	if nodeIsOnCooldown {
		return err
	}

	metricsValid, err := CheckIfMetricsValid(nodeID, &repos)
	if err != nil {
		return err
	}

	if metricsValid {
		err = repos.NodeRepo.AddNodeToActive(nodeID)
		if err != nil {
			log.Errorf("Unable to add node %s to active nodes, because of %v", nodeID, err)
		}
		log.Debugf("Node %s added to active nodes", nodeID)
	}

	return nil
}
