package payout

import (
	"errors"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	mocks "github.com/NodeFactoryIo/vedran/mocks/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func Test_GetStatsForPayout(t *testing.T) {
	now := time.Now()
	aDayAgo := now.Add(-24 * time.Hour)
	tests := []struct {
		name               string
		nodeId             string
		shouldRecordPayout bool
		// NodeRepo.GetAll
		nodeRepoGetAllReturns *[]models.Node
		nodeRepoGetAllError   error
		// NodeRepo.FindByID
		nodeRepoFindByIDReturns *models.Node
		nodeRepoFindByIDError   error
		// RecordRepo.FindSuccessfulRecordsInsideInterval
		recordRepoFindSuccessfulRecordsInsideIntervalReturns []models.Record
		recordRepoFindSuccessfulRecordsInsideIntervalError   error
		// DowntimeRepo.FindDowntimesInsideInterval
		downtimeRepoFindDowntimesInsideIntervalReturns []models.Downtime
		downtimeRepoFindDowntimesInsideIntervalError   error
		// PingRepo.CalculateDowntime
		pingRepoCalculateDowntimeReturnDuration time.Duration
		pingRepoCalculateDowntimeError          error
		// PayoutRepo.FindLatestPayout
		payoutRepoFindLatestPayoutReturns *models.Payout
		payoutRepoFindLatestPayoutError   error
		// PayoutRepo.Save
		payoutRepoSaveError      error
		payoutRepoSaveNumOfCalls int
		// GetStatsForPayout
		getStatsForPayoutReturns map[string]models.NodeStatsDetails
		getStatsForPayoutError   error
	}{
		{
			name:               "calculate stats and save payout record",
			nodeId:             "1",
			shouldRecordPayout: true,
			// NodeRepo.GetAll
			nodeRepoGetAllReturns: &[]models.Node{
				{
					ID: "1",
					PayoutAddress: "0xpayout-address-1",
				},
			},
			nodeRepoGetAllError: nil,
			// NodeRepo.FindByID
			nodeRepoFindByIDReturns: &models.Node{
				ID:            "1",
				PayoutAddress: "0xpayout-address-1",
			},
			nodeRepoFindByIDError: nil,
			// RecordRepo.FindSuccessfulRecordsInsideInterval
			recordRepoFindSuccessfulRecordsInsideIntervalReturns: []models.Record{
				{ID: 1, NodeId: "1", Status: "successful", Timestamp: now.Add(-12 * time.Hour)},
				{ID: 2, NodeId: "1", Status: "successful", Timestamp: now.Add(-10 * time.Hour)},
				{ID: 3, NodeId: "1", Status: "successful", Timestamp: now.Add(-8 * time.Hour)},
			},
			recordRepoFindSuccessfulRecordsInsideIntervalError: nil,
			// DowntimeRepo.FindDowntimesInsideInterval
			downtimeRepoFindDowntimesInsideIntervalReturns: nil,
			downtimeRepoFindDowntimesInsideIntervalError:   errors.New("not found"),
			// PingRepo.CalculateDowntime
			pingRepoCalculateDowntimeReturnDuration: 5 * time.Second,
			pingRepoCalculateDowntimeError:          nil,
			// PayoutRepo.FindLatestPayout
			payoutRepoFindLatestPayoutReturns: &models.Payout{
				Timestamp:      aDayAgo,
				PaymentDetails: nil,
			},
			payoutRepoFindLatestPayoutError: nil,
			// PayoutRepo.Save
			payoutRepoSaveError:      nil,
			payoutRepoSaveNumOfCalls: 1,
			// GetStatsForPayout
			getStatsForPayoutError: nil,
			getStatsForPayoutReturns: map[string]models.NodeStatsDetails{
				"0xpayout-address-1": {
					TotalPings:    17280,
					TotalRequests: 3,
				},
			},
		},
		{
			name:               "calculate stats and don't save payout record",
			nodeId:             "1",
			shouldRecordPayout: false,
			// NodeRepo.GetAll
			nodeRepoGetAllReturns: &[]models.Node{
				{
					ID: "1",
					PayoutAddress: "0xpayout-address-1",
				},
			},
			nodeRepoGetAllError: nil,
			// NodeRepo.FindByID
			nodeRepoFindByIDReturns: &models.Node{
				ID:            "1",
				PayoutAddress: "0xpayout-address-1",
			},
			nodeRepoFindByIDError: nil,
			// RecordRepo.FindSuccessfulRecordsInsideInterval
			recordRepoFindSuccessfulRecordsInsideIntervalReturns: []models.Record{
				{ID: 1, NodeId: "1", Status: "successful", Timestamp: now.Add(-12 * time.Hour)},
				{ID: 2, NodeId: "1", Status: "successful", Timestamp: now.Add(-10 * time.Hour)},
				{ID: 3, NodeId: "1", Status: "successful", Timestamp: now.Add(-8 * time.Hour)},
			},
			recordRepoFindSuccessfulRecordsInsideIntervalError: nil,
			// DowntimeRepo.FindDowntimesInsideInterval
			downtimeRepoFindDowntimesInsideIntervalReturns: nil,
			downtimeRepoFindDowntimesInsideIntervalError:   errors.New("not found"),
			// PingRepo.CalculateDowntime
			pingRepoCalculateDowntimeReturnDuration: 5 * time.Second,
			pingRepoCalculateDowntimeError:          nil,
			// PayoutRepo.FindLatestPayout
			payoutRepoFindLatestPayoutReturns: &models.Payout{
				Timestamp:      aDayAgo,
				PaymentDetails: nil,
			},
			payoutRepoFindLatestPayoutError: nil,
			// PayoutRepo.Save
			payoutRepoSaveError:      nil,
			payoutRepoSaveNumOfCalls: 0,
			// GetStatsForPayout
			getStatsForPayoutError: nil,
			getStatsForPayoutReturns: map[string]models.NodeStatsDetails{
				"0xpayout-address-1": {
					TotalPings:    17280,
					TotalRequests: 3,
				},
			},
		},
		{
			name:               "fail on fetching from database",
			nodeId:             "1",
			shouldRecordPayout: false,
			// PayoutRepo.FindLatestPayout
			payoutRepoFindLatestPayoutError: errors.New("db error"),
			// GetStatsForPayout
			getStatsForPayoutError:   errors.New("db error"),
			getStatsForPayoutReturns: nil,
		},
		{
			name:               "fail on saving to database",
			nodeId:             "1",
			shouldRecordPayout: true,
			// NodeRepo.GetAll
			nodeRepoGetAllReturns: &[]models.Node{
				{
					ID: "1",
					PayoutAddress: "0xpayout-address-1",
				},
			},
			nodeRepoGetAllError: nil,
			// NodeRepo.FindByID
			nodeRepoFindByIDReturns: &models.Node{
				ID:            "1",
				PayoutAddress: "0xpayout-address-1",
			},
			nodeRepoFindByIDError: nil,
			// RecordRepo.FindSuccessfulRecordsInsideInterval
			recordRepoFindSuccessfulRecordsInsideIntervalReturns: []models.Record{
				{ID: 1, NodeId: "1", Status: "successful", Timestamp: now.Add(-12 * time.Hour)},
				{ID: 2, NodeId: "1", Status: "successful", Timestamp: now.Add(-10 * time.Hour)},
				{ID: 3, NodeId: "1", Status: "successful", Timestamp: now.Add(-8 * time.Hour)},
			},
			recordRepoFindSuccessfulRecordsInsideIntervalError: nil,
			// DowntimeRepo.FindDowntimesInsideInterval
			downtimeRepoFindDowntimesInsideIntervalReturns: nil,
			downtimeRepoFindDowntimesInsideIntervalError:   errors.New("not found"),
			// PingRepo.CalculateDowntime
			pingRepoCalculateDowntimeReturnDuration: 5 * time.Second,
			pingRepoCalculateDowntimeError:          nil,
			// PayoutRepo.FindLatestPayout
			payoutRepoFindLatestPayoutReturns: &models.Payout{
				Timestamp:      aDayAgo,
				PaymentDetails: nil,
			},
			payoutRepoFindLatestPayoutError: nil,
			// PayoutRepo.Save
			payoutRepoSaveError:      errors.New("db error"),
			payoutRepoSaveNumOfCalls: 1,
			// GetStatsForPayout
			getStatsForPayoutError:   errors.New("db error"),
			getStatsForPayoutReturns: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			nodeRepoMock := mocks.NodeRepository{}
			nodeRepoMock.On("GetAll").Return(
				test.nodeRepoGetAllReturns, test.nodeRepoGetAllError,
			)
			nodeRepoMock.On("FindByID", test.nodeId).Return(
				test.nodeRepoFindByIDReturns, test.nodeRepoFindByIDError,
			)
			recordRepoMock := mocks.RecordRepository{}
			recordRepoMock.On("FindSuccessfulRecordsInsideInterval",
				test.nodeId, mock.Anything, mock.Anything,
			).Return(
				test.recordRepoFindSuccessfulRecordsInsideIntervalReturns,
				test.recordRepoFindSuccessfulRecordsInsideIntervalError,
			)
			pingRepoMock := mocks.PingRepository{}
			pingRepoMock.On("CalculateDowntime",
				test.nodeId, mock.Anything,
			).Return(
				time.Now(),
				test.pingRepoCalculateDowntimeReturnDuration,
				test.pingRepoCalculateDowntimeError,
			)
			downtimeRepoMock := mocks.DowntimeRepository{}
			downtimeRepoMock.On("FindDowntimesInsideInterval",
				test.nodeId, mock.Anything, mock.Anything,
			).Return(
				test.downtimeRepoFindDowntimesInsideIntervalReturns,
				test.downtimeRepoFindDowntimesInsideIntervalError,
			)
			payoutRepoMock := mocks.PayoutRepository{}
			payoutRepoMock.On("FindLatestPayout").Return(
				test.payoutRepoFindLatestPayoutReturns,
				test.payoutRepoFindLatestPayoutError,
			)
			payoutRepoMock.On("Save", mock.Anything).Return(test.payoutRepoSaveError)

			repos := repositories.Repos{
				NodeRepo:     &nodeRepoMock,
				PingRepo:     &pingRepoMock,
				RecordRepo:   &recordRepoMock,
				DowntimeRepo: &downtimeRepoMock,
				PayoutRepo:   &payoutRepoMock,
			}

			_, statsForPayout, err := GetStatsForPayout(repos, now, test.shouldRecordPayout)
			assert.Equal(t, test.getStatsForPayoutReturns, statsForPayout)
			assert.Equal(t, test.getStatsForPayoutError, err)
			// check if payout saved
			payoutRepoMock.AssertNumberOfCalls(t, "Save", test.payoutRepoSaveNumOfCalls)
		})
	}
}
