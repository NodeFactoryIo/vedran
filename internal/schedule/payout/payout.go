package schedulepayout

import (
	"errors"
	"time"

	"github.com/NodeFactoryIo/vedran/internal/configuration"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"github.com/NodeFactoryIo/vedran/internal/script"
	"github.com/NodeFactoryIo/vedran/internal/ui"
	log "github.com/sirupsen/logrus"
)

// StartScheduledPayout checks every 24 hours how many days have passed since last payout.
// If number of passed days is equal or bigger than defined interval in configuration, start automatic payout
func StartScheduledPayout(configuration configuration.PayoutConfiguration, privateKey string, repos repositories.Repos) {
	ticker := time.NewTicker(time.Hour * 24)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				checkForPayout(privateKey, configuration, repos)
			}
		}
	}()
}

// GetNextPayoutDate returns date of next scheduled payout or error if payout disabled
func GetNextPayoutDate(configuration *configuration.PayoutConfiguration, repos repositories.Repos) (time.Time, error) {
	if configuration == nil {
		return time.Now(), errors.New("Schedule payout not configured")
	}

	latestPayout, err := repos.PayoutRepo.FindLatestPayout()
	if err != nil {
		log.Errorf("Unable to calculate last payout because of: %v", err)
		return time.Now(), err
	}

	return latestPayout.Timestamp.AddDate(0, 0, configuration.PayoutNumberOfDays), nil
}

func checkForPayout(
	privateKey string,
	configuration configuration.PayoutConfiguration,
	repos repositories.Repos,
) {
	daysSinceLastPayout, lastPayoutTimestamp, err := numOfDaysSinceLastPayout(repos)
	if err != nil {
		log.Error("Unable to calculate number of days since last payout", err)
		return
	}

	if daysSinceLastPayout >= configuration.PayoutNumberOfDays {
		go startPayout(privateKey, configuration)
	} else {
		log.Infof(
			"Last payout was %s, next payout will be in %d days",
			lastPayoutTimestamp.Format("2006-January-02"),
			configuration.PayoutNumberOfDays-daysSinceLastPayout,
		)
	}
}

func startPayout(privateKey string, configuration configuration.PayoutConfiguration) {
	log.Info("Starting automatic payout...")
	transactionDetails, err := script.ExecutePayout(
		privateKey,
		configuration.PayoutTotalReward,
		configuration.LbFeeAddress,
		configuration.LbURL,
	)
	if transactionDetails != nil {
		// display even if only part of transactions executed
		ui.DisplayTransactionsStatus(transactionDetails)
	}
	if err != nil {
		log.Errorf("Unable to execute payout, because of: %v", err)
		return
	} else {
		log.Info("Payout execution finished")
	}
}

func numOfDaysSinceLastPayout(repos repositories.Repos) (int, *time.Time, error) {
	latestPayout, err := repos.PayoutRepo.FindLatestPayout()
	if err != nil {
		return 0, nil, err
	}
	daysSinceLastPayout := time.Since(latestPayout.Timestamp) / (24 * time.Hour)
	return int(daysSinceLastPayout), &latestPayout.Timestamp, nil
}
