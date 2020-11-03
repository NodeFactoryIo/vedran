package stats

import (
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"time"
)

var nowFunc = time.Now

// GetIntervalFromLastPayout returns interval from last recorded payout until now as (intervalStart, intervalEnd, err)
func GetIntervalFromLastPayout(repos repositories.Repos) (*time.Time, *time.Time, error) {
	latestPayout, err := repos.PayoutRepo.FindLatestPayout()
	if err != nil {
		return nil, nil, err
	}

	intervalStart := latestPayout.Timestamp
	intervalEnd := nowFunc()

	return &intervalStart, &intervalEnd, err
}