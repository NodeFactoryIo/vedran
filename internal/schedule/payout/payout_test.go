package schedulepayout

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/NodeFactoryIo/vedran/internal/configuration"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	mocks "github.com/NodeFactoryIo/vedran/mocks/repositories"
	"github.com/stretchr/testify/assert"
)

var (
	timestampSince20days = time.Now().Add(-20 * (24 * time.Hour))
	timestampSince2days  = time.Now().Add(-2 * (24 * time.Hour))
	timestampSince2hours = time.Now().Add(-2 * time.Hour)
)

func Test_numOfDaysSinceLastPayout(t *testing.T) {
	tests := []struct {
		name                              string
		latestPayout                      *models.Payout
		latestPayoutError                 error
		numOfDaysSinceLastPayoutError     error
		numOfDaysSinceLastPayoutNumOfDays int
		numOfDaysSinceLastPayoutTimestamp *time.Time
	}{
		{
			name: "last payout before 20 days",
			latestPayout: &models.Payout{
				ID:             1,
				Timestamp:      timestampSince20days,
				PaymentDetails: nil,
			},
			latestPayoutError:                 nil,
			numOfDaysSinceLastPayoutError:     nil,
			numOfDaysSinceLastPayoutNumOfDays: 20,
			numOfDaysSinceLastPayoutTimestamp: &timestampSince20days,
		},
		{
			name: "last payout before 2 days",
			latestPayout: &models.Payout{
				ID:             1,
				Timestamp:      timestampSince2days,
				PaymentDetails: nil,
			},
			latestPayoutError:                 nil,
			numOfDaysSinceLastPayoutError:     nil,
			numOfDaysSinceLastPayoutNumOfDays: 2,
			numOfDaysSinceLastPayoutTimestamp: &timestampSince2days,
		},
		{
			name: "last payout before 2 hours",
			latestPayout: &models.Payout{
				ID:             1,
				Timestamp:      timestampSince2hours,
				PaymentDetails: nil,
			},
			latestPayoutError:                 nil,
			numOfDaysSinceLastPayoutError:     nil,
			numOfDaysSinceLastPayoutNumOfDays: 0,
			numOfDaysSinceLastPayoutTimestamp: &timestampSince2hours,
		},
		{
			name:                              "error on latest payout",
			latestPayout:                      nil,
			latestPayoutError:                 errors.New("db error"),
			numOfDaysSinceLastPayoutError:     errors.New("db error"),
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

func TestGetNextPayoutDate(t *testing.T) {
	type args struct {
		configuration *configuration.PayoutConfiguration
	}
	tests := []struct {
		name              string
		args              args
		want              time.Time
		wantErr           bool
		latestPayout      *models.Payout
		latestPayoutError error
	}{
		{
			name: "Returns error if payout configuration not defined",
			latestPayout: &models.Payout{
				ID:             1,
				Timestamp:      timestampSince20days,
				PaymentDetails: nil,
			},
			latestPayoutError: nil,
			args:              args{configuration: nil},
			want:              time.Now(),
			wantErr:           true,
		},
		{
			name: "Returns error if calculating lastPayoutTimestamp fails",
			latestPayout: &models.Payout{
				ID:             1,
				Timestamp:      timestampSince20days,
				PaymentDetails: nil,
			},
			latestPayoutError: errors.New("Error"),
			args: args{configuration: &configuration.PayoutConfiguration{
				PayoutNumberOfDays: 25,
			}},
			want:    time.Now(),
			wantErr: true,
		},
		{
			name: "Returns next payout date if configuration valid and lastPayout exists",
			latestPayout: &models.Payout{
				ID:             1,
				Timestamp:      timestampSince20days,
				PaymentDetails: nil,
			},
			latestPayoutError: nil,
			args: args{configuration: &configuration.PayoutConfiguration{
				PayoutNumberOfDays: 25,
			}},
			want:    timestampSince20days.AddDate(0, 0, 25),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payoutRepoMock := mocks.PayoutRepository{}
			payoutRepoMock.On("FindLatestPayout").Return(
				tt.latestPayout, tt.latestPayoutError,
			)
			repos := repositories.Repos{
				PayoutRepo: &payoutRepoMock,
			}

			got, err := GetNextPayoutDate(tt.args.configuration, repos)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetNextPayoutDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr == false && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNextPayoutDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
