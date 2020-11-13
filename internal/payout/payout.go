package payout

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"github.com/NodeFactoryIo/vedran/internal/stats"
	log "github.com/sirupsen/logrus"
	"time"
)

type NodePayoutDetails struct {
	Stats         models.NodeStatsDetails `json:"stats"`
	PayoutAddress string                  `json:"payout_address"`
}

func GetStatsForPayout(
	repos repositories.Repos,
	intervalEnd time.Time,
	recordPayout bool,
) (map[string]NodePayoutDetails, error) {

	statistics, err := stats.CalculateStatisticsFromLastPayout(repos, intervalEnd)
	if err != nil {
		return nil, err
	}

	payoutStatistics := make(map[string]NodePayoutDetails, len(statistics))
	for nodeId, statsDetails := range statistics {
		node, _ := repos.NodeRepo.FindByID(nodeId)
		payoutStatistics[nodeId] = NodePayoutDetails{
			Stats:         statsDetails,
			PayoutAddress: node.PayoutAddress,
		}
	}

	if recordPayout {
		err = repos.PayoutRepo.Save(&models.Payout{
			Timestamp:      intervalEnd,
			PaymentDetails: statistics,
		})
		if err != nil {
			log.Errorf("Unable to save payout information To database, because of: %v", err)
			return nil, err
		}
	}

	return payoutStatistics, nil
}
