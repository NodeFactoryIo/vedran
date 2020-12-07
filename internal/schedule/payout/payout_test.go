package schedulepayout

import (
	"errors"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	mocks "github.com/NodeFactoryIo/vedran/mocks/repositories"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	timestampSince20days = time.Now().Add(-20 * (24 * time.Hour))
	timestampSince2days = time.Now().Add(-2 * (24 * time.Hour))
	timestampSince2hours = time.Now().Add(-2 * time.Hour)
)

func Test_numOfDaysSinceLastPayout(t *testing.T) {
	tests := []struct {
		name string
		latestPayout *models.Payout
		latestPayoutError error
		numOfDaysSinceLastPayoutError error
		numOfDaysSinceLastPayoutNumOfDays int
		numOfDaysSinceLastPayoutTimestamp *time.Time
	}{
		{
			name: "last payout before 20 days",
			latestPayout: &models.Payout{
				ID:             "",
				Timestamp:      timestampSince20days,
				PaymentDetails: nil,
			},
			latestPayoutError: nil,
			numOfDaysSinceLastPayoutError: nil,
			numOfDaysSinceLastPayoutNumOfDays: 20,
			numOfDaysSinceLastPayoutTimestamp: &timestampSince20days,
		},
		{
			name: "last payout before 2 days",
			latestPayout: &models.Payout{
				ID:             "",
				Timestamp:      timestampSince2days,
				PaymentDetails: nil,
			},
			latestPayoutError: nil,
			numOfDaysSinceLastPayoutError: nil,
			numOfDaysSinceLastPayoutNumOfDays: 2,
			numOfDaysSinceLastPayoutTimestamp: &timestampSince2days,
		},
		{
			name: "last payout before 2 hours",
			latestPayout: &models.Payout{
				ID:             "",
				Timestamp:      timestampSince2hours,
				PaymentDetails: nil,
			},
			latestPayoutError: nil,
			numOfDaysSinceLastPayoutError: nil,
			numOfDaysSinceLastPayoutNumOfDays: 0,
			numOfDaysSinceLastPayoutTimestamp: &timestampSince2hours,
		},
		{
			name: "error on latest payout",
			latestPayout: nil,
			latestPayoutError: errors.New("db error"),
			numOfDaysSinceLastPayoutError: errors.New("db error"),
			numOfDaysSinceLastPayoutNumOfDays: 0,
			numOfDaysSinceLastPayoutTimestamp: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payoutRepoMock := mocks.PayoutRepository{}
			payoutRepoMock.On("FindLatestPayout").Return(
				test.latestPayout, test.latestPayoutError,
			)
			repos := repositories.Repos{
				PayoutRepo: &payoutRepoMock,
			}

			daysSinceLastPayout, lastPayoutTimestamp, err := numOfDaysSinceLastPayout(repos)
			assert.Equal(t, test.latestPayoutError, err)
			assert.Equal(t, test.numOfDaysSinceLastPayoutNumOfDays, daysSinceLastPayout)
			assert.Equal(t, test.numOfDaysSinceLastPayoutTimestamp, lastPayoutTimestamp)
		})
	}
}

