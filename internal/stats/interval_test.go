package stats

import (
	"errors"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	mocks "github.com/NodeFactoryIo/vedran/mocks/repositories"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_GetIntervalFromLastPayout(t *testing.T) {
	hourAgo := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name string
		payoutRepoFindLatestPayoutReturns *models.Payout
		payoutRepoFindLatestPayoutError error
		getIntervalFromLastPayoutIntervalStart *time.Time
		getIntervalFromLastPayoutError error
	}{
		{
			name: "calculate interval from existing last payment to now",
			payoutRepoFindLatestPayoutReturns: &models.Payout{
				Timestamp:      hourAgo,
				PaymentDetails: nil,
			},
			payoutRepoFindLatestPayoutError: nil,
			getIntervalFromLastPayoutIntervalStart: &hourAgo,
			getIntervalFromLastPayoutError: nil,
		},
		{
			name: "calculate interval from existing last payment to now",
			payoutRepoFindLatestPayoutReturns: nil,
			payoutRepoFindLatestPayoutError: errors.New("db error"),
			getIntervalFromLastPayoutIntervalStart: nil,
			getIntervalFromLastPayoutError: errors.New("db error"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payoutRepoMock := mocks.PayoutRepository{}
			payoutRepoMock.On("FindLatestPayout").Return(
				test.payoutRepoFindLatestPayoutReturns,
				test.payoutRepoFindLatestPayoutError,
			)
			repos := repositories.Repos{
				PayoutRepo:   &payoutRepoMock,
			}

			intervalStart, err := GetIntervalFromLastPayout(repos)
			assert.Equal(t, test.getIntervalFromLastPayoutIntervalStart, intervalStart)
			assert.Equal(t, test.getIntervalFromLastPayoutError, err)
		})
	}
}
