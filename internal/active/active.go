package active

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	IntervalFromLastPing = 10 * time.Second
	AllowedBlocksBehind = 10
)

// CheckIfNodeActive checks if nodes last recorded ping is in last IntervalFromLastPing and if nodes last recorded
// BestBlockHeight and FinalizedBlockHeight are lagging more than AllowedBlocksBehind blocks
func CheckIfNodeActive(node models.Node, repos *repositories.Repos) (bool, error) {
	lastPing, err := repos.PingRepo.FindByNodeID(node.ID)
	if err != nil {
		return false, err
	}

	// more than 10 seconds passed from last ping
	if lastPing.Timestamp.Add(IntervalFromLastPing).Before(time.Now()) {
		log.Debugf("Node %s not active as last ping was at %v", node.ID, lastPing.Timestamp)
		return false, nil
	}

	// node's latest and best block lag behind the best in the pool by more than 10 blocks
	metrics, err := repos.MetricsRepo.FindByID(node.ID)
	if err != nil {
		return false, err
	}
	latestBlockMetrics, err := repos.MetricsRepo.GetLatestBlockMetrics()
	if err != nil {
		return false, err
	}
	if metrics.BestBlockHeight <= (latestBlockMetrics.BestBlockHeight - AllowedBlocksBehind) ||
		metrics.FinalizedBlockHeight <= (latestBlockMetrics.FinalizedBlockHeight - AllowedBlocksBehind) {
		log.Debugf("Node %s not active as metrics check failed. BestBlockHeight: %d, FinalizedBlockHeight: %d",
			node.ID, metrics.BestBlockHeight, metrics.FinalizedBlockHeight)
		return false, nil
	}
	
	return true, nil
}
