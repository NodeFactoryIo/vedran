package stats

import (
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"time"
)

// GetIntervalFromLastPayout returns interval from last recorded payout until now as (intervalStart, intervalEnd, err)
func GetIntervalFromLastPayout(repos repositories.Repos) (*time.Time, error) {
	latestPayout, err := repos.PayoutRepo.FindLatestPayout()
	if err != nil {
		return nil, err
	}

	intervalStart := latestPayout.Timestamp
	return &intervalStart, err
}