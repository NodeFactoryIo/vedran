package payout

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"github.com/NodeFactoryIo/vedran/internal/stats"
	log "github.com/sirupsen/logrus"
	"time"
)

func GetStatsForPayout(
	repos repositories.Repos,
	intervalEnd time.Time,
	recordPayout bool,
) (map[string]models.NodeStatsDetails, error) {
	statistics, err := stats.CalculateStatisticsFromLastPayout(repos, intervalEnd)
	if err != nil {
		return nil, err
	}

	if recordPayout {
		err = repos.PayoutRepo.Save(&models.Payout{
			Timestamp:      intervalEnd,
			PaymentDetails: statistics,
		})
		if err != nil {
			log.Errorf("Unable to save payout information to database, because of: %v", err)
			return nil, err
		}
	}

	return statistics, nil
}
