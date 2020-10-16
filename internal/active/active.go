package active

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"time"
)

const IntervalFromLastPing = 10 * time.Second
const AllowedBlocksBehind = 10

// CheckIfNodeActive checks if nodes last recorded ping is in last IntervalFromLastPing and if nodes last recorded
// BestBlockHeight and FinalizedBlockHeight are lagging more than AllowedBlocksBehind blocks
func CheckIfNodeActive(node models.Node, repos *repositories.Repos) (bool, error) {
	lastPing, err := repos.PingRepo.FindByNodeID(node.ID)
	if err != nil {
		return false, err
	}

	// more than 10 seconds passed from last ping
	if lastPing.Timestamp.Add(IntervalFromLastPing).Before(time.Now()) {
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
		return false, nil
	}
	
	return true, nil
}
