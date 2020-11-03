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
	now := time.Now()
	hourAgo := now.Add(-24 * time.Hour)
	nowFunc = func() time.Time {
		return now
	}

	tests := []struct {
		name string
		payoutRepoFindLatestPayoutReturns *models.Payout
		payoutRepoFindLatestPayoutError error
		getIntervalFromLastPayoutIntervalStart *time.Time
		getIntervalFromLastPayoutIntervalEnd *time.Time
		getIntervalFromLastPayoutError error
	}{
		{
			name: "calculate interval from existing last payment to now",
			payoutRepoFindLatestPayoutReturns: &models.Payout{
				Timestamp:      now.Add(-24 * time.Hour),
				PaymentDetails: nil,
			},
			payoutRepoFindLatestPayoutError: nil,
			getIntervalFromLastPayoutIntervalStart: &hourAgo,
			getIntervalFromLastPayoutIntervalEnd: &now,
			getIntervalFromLastPayoutError: nil,
		},
		{
			name: "calculate interval from existing last payment to now",
			payoutRepoFindLatestPayoutReturns: nil,
			payoutRepoFindLatestPayoutError: errors.New("db error"),
			getIntervalFromLastPayoutIntervalStart: nil,
			getIntervalFromLastPayoutIntervalEnd: nil,
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

			intervalStart, intervalEnd, err := GetIntervalFromLastPayout(repos)
			assert.Equal(t, test.getIntervalFromLastPayoutIntervalStart, intervalStart)
			assert.Equal(t, test.getIntervalFromLastPayoutIntervalEnd, intervalEnd)
			assert.Equal(t, test.getIntervalFromLastPayoutError, err)
		})
	}

	nowFunc = time.Now
}
