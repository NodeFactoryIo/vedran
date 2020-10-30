package stats

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	log "github.com/sirupsen/logrus"
	"time"
)

// CalculateStatisticsForInterval TODO
func CalculateStatisticsForInterval(
	repos repositories.Repos,
	intervalStart time.Time,
	intervalEnd time.Time,
) (map[string]models.NodeStatsDetails, error) {

	allNodes, err := repos.NodeRepo.GetAll()
	if err != nil {
		if err.Error() == "not found" {
			log.Debugf("Unable to calculate statistics if there isn't any saved nodes")
		}
		return nil, err
	}

	var allNodesStats = make(map[string]models.NodeStatsDetails)
	for _, node := range *allNodes {
		nodeStats, err := CalculateNodeStatisticsForInterval(repos, node.ID, intervalStart, intervalEnd)
		if err != nil {
			return nil, err
		}
		allNodesStats[node.ID] = *nodeStats
	}

	return allNodesStats, nil
}

// CalculateNodeStatisticsForInterval TODO
func CalculateNodeStatisticsForInterval(
	repos repositories.Repos,
	nodeId string,
	intervalStart time.Time,
	intervalEnd time.Time,
) (*models.NodeStatsDetails, error) {
	recordsInInterval, err := repos.RecordRepo.FindSuccessfulRecordsInsideInterval(nodeId, intervalStart, intervalEnd)
	if err != nil {
		if err.Error() == "not found" {
			recordsInInterval = []models.Record{}
		} else {
			return nil, err
		}
	}

	totalPings, err := CalculateTotalPingsForNode(repos, nodeId, intervalStart, intervalEnd)
	if err != nil {
		log.Errorf("Unable to calculate total number of pings for node %s, because %v", nodeId, err)
		return nil, err
	}

	return &models.NodeStatsDetails{
		TotalPings:    totalPings,
		TotalRequests: float64(len(recordsInInterval)),
	}, nil
}
