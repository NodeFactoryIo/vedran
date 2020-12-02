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

func Test_CalculateNodeStatisticsFromLastPayout(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		nodeID        string
		intervalStart time.Time
		intervalEnd   time.Time
		// RecordRepo.FindByNodeID
		recordRepoFindSuccessfulRecordsInsideIntervalReturns    []models.Record
		recordRepoFindSuccessfulRecordsInsideIntervalError      error
		recordRepoFindSuccessfulRecordsInsideIntervalNumOfCalls int
		// DowntimeRepo.FindByNodeID
		downtimeRepoFindDowntimesInsideIntervalReturns    []models.Downtime
		downtimeRepoFindDowntimesInsideIntervalError      error
		downtimeRepoFindDowntimesInsideIntervalNumOfCalls int
		// PingRepo.CalculateDowntime
		pingRepoCalculateDowntimeReturnDuration time.Duration
		pingRepoCalculateDowntimeError          error
		pingRepoCalculateDowntimeNumOfCalls     int
		// PayoutRepo.FindLatestPayout
		payoutRepoFindLatestPayoutReturns    *models.Payout
		payoutRepoFindLatestPayoutError      error
		payoutRepoFindLatestPayoutNumOfCalls int
		// CalculateNodeStatisticsForInterval
		calculateNodeStatisticsFromLastPayoutReturns *models.NodeStatsDetails
		calculateNodeStatisticsFromLastPayoutError   error
	}{
		{
			name:          "valid statistics with no records and no downtime",
			nodeID:        "1",
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// RecordRepo.FindByNodeID
			recordRepoFindSuccessfulRecordsInsideIntervalReturns:    nil,
			recordRepoFindSuccessfulRecordsInsideIntervalError:      errors.New("not found"),
			recordRepoFindSuccessfulRecordsInsideIntervalNumOfCalls: 1,
			// DowntimeRepo.FindByNodeID
			downtimeRepoFindDowntimesInsideIntervalReturns:    nil,
			downtimeRepoFindDowntimesInsideIntervalError:      errors.New("not found"),
			downtimeRepoFindDowntimesInsideIntervalNumOfCalls: 1,
			// PingRepo.CalculateDowntime
			pingRepoCalculateDowntimeReturnDuration: 5 * time.Second,
			pingRepoCalculateDowntimeError:          nil,
			pingRepoCalculateDowntimeNumOfCalls:     1,
			// PayoutRepo.FindLatestPayout
			payoutRepoFindLatestPayoutReturns: &models.Payout{
				Timestamp:      now.Add(-24 * time.Hour),
				PaymentDetails: nil,
			},
			payoutRepoFindLatestPayoutError:      nil,
			payoutRepoFindLatestPayoutNumOfCalls: 1,
			// CalculateNodeStatisticsForInterval
			calculateNodeStatisticsFromLastPayoutReturns: &models.NodeStatsDetails{
				TotalPings:    17280, // no downtime - max number of pings
				TotalRequests: 0,
			},
			calculateNodeStatisticsFromLastPayoutError: nil,
		},
		{
			name:          "error on fetching latest payout",
			nodeID:        "1",
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// PayoutRepo.FindLatestPayout
			payoutRepoFindLatestPayoutReturns:    nil,
			payoutRepoFindLatestPayoutError:      errors.New("db error"),
			payoutRepoFindLatestPayoutNumOfCalls: 1,
			// CalculateNodeStatisticsForInterval
			calculateNodeStatisticsFromLastPayoutReturns: nil,
			calculateNodeStatisticsFromLastPayoutError:   errors.New("db error"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create mock controller
			recordRepoMock := mocks.RecordRepository{}
			recordRepoMock.On("FindSuccessfulRecordsInsideInterval",
				test.nodeID, test.intervalStart, test.intervalEnd,
			).Return(
				test.recordRepoFindSuccessfulRecordsInsideIntervalReturns,
				test.recordRepoFindSuccessfulRecordsInsideIntervalError,
			)
			downtimeRepoMock := mocks.DowntimeRepository{}
			downtimeRepoMock.On("FindDowntimesInsideInterval",
				test.nodeID, test.intervalStart, test.intervalEnd,
			).Return(
				test.downtimeRepoFindDowntimesInsideIntervalReturns,
				test.downtimeRepoFindDowntimesInsideIntervalError,
			)
			pingRepoMock := mocks.PingRepository{}
			pingRepoMock.On("CalculateDowntime",
				test.nodeID, test.intervalEnd,
			).Return(
				time.Now(),
				test.pingRepoCalculateDowntimeReturnDuration,
				test.pingRepoCalculateDowntimeError,
			)
			payoutRepoMock := mocks.PayoutRepository{}
			payoutRepoMock.On("FindLatestPayout").Return(
				test.payoutRepoFindLatestPayoutReturns,
				test.payoutRepoFindLatestPayoutError,
			)
			repos := repositories.Repos{
				PingRepo:     &pingRepoMock,
				RecordRepo:   &recordRepoMock,
				DowntimeRepo: &downtimeRepoMock,
				PayoutRepo:   &payoutRepoMock,
			}

			statisticsForPayout, err := CalculateNodeStatisticsFromLastPayout(repos, test.nodeID, test.intervalEnd)

			assert.Equal(t, test.calculateNodeStatisticsFromLastPayoutError, err)
			assert.Equal(t, test.calculateNodeStatisticsFromLastPayoutReturns, statisticsForPayout)

			recordRepoMock.AssertNumberOfCalls(t,
				"FindSuccessfulRecordsInsideInterval",
				test.recordRepoFindSuccessfulRecordsInsideIntervalNumOfCalls,
			)
			downtimeRepoMock.AssertNumberOfCalls(t,
				"FindDowntimesInsideInterval",
				test.downtimeRepoFindDowntimesInsideIntervalNumOfCalls,
			)
			pingRepoMock.AssertNumberOfCalls(t,
				"CalculateDowntime",
				test.pingRepoCalculateDowntimeNumOfCalls,
			)
			payoutRepoMock.AssertNumberOfCalls(t,
				"FindLatestPayout",
				test.payoutRepoFindLatestPayoutNumOfCalls,
			)
		})
	}
}

func Test_CalculateStatisticsFromLastPayout(t *testing.T) {
	now := time.Now()

	testNode := models.Node{
		ID: "1",
		PayoutAddress: "0xpayout-address-1",
	}

	tests := []struct {
		name          string
		intervalStart time.Time
		intervalEnd   time.Time
		// NodeRepo.GetAll
		nodeRepoGetAllReturns    *[]models.Node
		nodeRepoGetAllError      error
		nodeRepoGetAllNumOfCalls int
		// RecordRepo.FindByNodeID
		recordRepoFindSuccessfulRecordsInsideIntervalReturns    []models.Record
		recordRepoFindSuccessfulRecordsInsideIntervalError      error
		recordRepoFindSuccessfulRecordsInsideIntervalNumOfCalls int
		// DowntimeRepo.FindByNodeID
		downtimeRepoFindDowntimesInsideIntervalReturns    []models.Downtime
		downtimeRepoFindDowntimesInsideIntervalError      error
		downtimeRepoFindDowntimesInsideIntervalNumOfCalls int
		// PingRepo.CalculateDowntime
		pingRepoCalculateDowntimeReturnDuration time.Duration
		pingRepoCalculateDowntimeError          error
		pingRepoCalculateDowntimeNumOfCalls     int
		// PayoutRepo.FindLatestPayout
		payoutRepoFindLatestPayoutReturns    *models.Payout
		payoutRepoFindLatestPayoutError      error
		payoutRepoFindLatestPayoutNumOfCalls int
		// CalculateNodeStatisticsForInterval
		calculateStatisticsFromLastPayoutReturns map[string]models.NodeStatsDetails
		calculateStatisticsFromLastPayoutError   error
	}{
		{
			name:   "valid statistics with multiple records and no downtime",
			// interval of 24 hours
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// NodeRepo.GetAll
			nodeRepoGetAllReturns: &[]models.Node{
				testNode,
			},
			nodeRepoGetAllError:      nil,
			nodeRepoGetAllNumOfCalls: 1,
			// RecordRepo.FindByNodeID
			recordRepoFindSuccessfulRecordsInsideIntervalReturns:    nil,
			recordRepoFindSuccessfulRecordsInsideIntervalError:      errors.New("not found"),
			recordRepoFindSuccessfulRecordsInsideIntervalNumOfCalls: 1,
			// DowntimeRepo.FindByNodeID
			downtimeRepoFindDowntimesInsideIntervalReturns:    nil,
			downtimeRepoFindDowntimesInsideIntervalError:      errors.New("not found"),
			downtimeRepoFindDowntimesInsideIntervalNumOfCalls: 1,
			// PingRepo.CalculateDowntime
			pingRepoCalculateDowntimeReturnDuration: 5 * time.Second,
			pingRepoCalculateDowntimeError:          nil,
			pingRepoCalculateDowntimeNumOfCalls:     1,
			// PayoutRepo.FindLatestPayout
			payoutRepoFindLatestPayoutReturns: &models.Payout{
				Timestamp:      now.Add(-24 * time.Hour),
				PaymentDetails: nil,
			},
			payoutRepoFindLatestPayoutError:      nil,
			payoutRepoFindLatestPayoutNumOfCalls: 1,
			// CalculateNodeStatisticsForInterval
			calculateStatisticsFromLastPayoutReturns: map[string]models.NodeStatsDetails{
				testNode.PayoutAddress: {
					TotalPings:    17280, // no downtime - max number of pings
					TotalRequests: 0,
				},
			},
			calculateStatisticsFromLastPayoutError: nil,
		},
		{
			name:   "error on fetching latest payout",
			// interval of 24 hours
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// PayoutRepo.FindLatestPayout
			payoutRepoFindLatestPayoutReturns:    nil,
			payoutRepoFindLatestPayoutError:      errors.New("db error"),
			payoutRepoFindLatestPayoutNumOfCalls: 1,
			// CalculateNodeStatisticsForInterval
			calculateStatisticsFromLastPayoutReturns: nil,
			calculateStatisticsFromLastPayoutError:   errors.New("db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create mock controller
			nodeRepoMock := mocks.NodeRepository{}
			nodeRepoMock.On("GetAll").Return(
				test.nodeRepoGetAllReturns, test.nodeRepoGetAllError,
			)
			recordRepoMock := mocks.RecordRepository{}
			recordRepoMock.On("FindSuccessfulRecordsInsideInterval",
				testNode.ID, test.intervalStart, test.intervalEnd,
			).Return(
				test.recordRepoFindSuccessfulRecordsInsideIntervalReturns,
				test.recordRepoFindSuccessfulRecordsInsideIntervalError,
			)
			downtimeRepoMock := mocks.DowntimeRepository{}
			downtimeRepoMock.On("FindDowntimesInsideInterval",
				testNode.ID, test.intervalStart, test.intervalEnd,
			).Return(
				test.downtimeRepoFindDowntimesInsideIntervalReturns,
				test.downtimeRepoFindDowntimesInsideIntervalError,
			)
			pingRepoMock := mocks.PingRepository{}
			pingRepoMock.On("CalculateDowntime",
				testNode.ID, test.intervalEnd,
			).Return(
				time.Now(),
				test.pingRepoCalculateDowntimeReturnDuration,
				test.pingRepoCalculateDowntimeError,
			)
			payoutRepoMock := mocks.PayoutRepository{}
			payoutRepoMock.On("FindLatestPayout").Return(
				test.payoutRepoFindLatestPayoutReturns,
				test.payoutRepoFindLatestPayoutError,
			)
			repos := repositories.Repos{
				PingRepo:     &pingRepoMock,
				RecordRepo:   &recordRepoMock,
				DowntimeRepo: &downtimeRepoMock,
				NodeRepo:     &nodeRepoMock,
				PayoutRepo:   &payoutRepoMock,
			}

			statisticsForPayout, err := CalculateStatisticsFromLastPayout(repos, test.intervalEnd)

			assert.Equal(t, test.calculateStatisticsFromLastPayoutError, err)
			assert.Equal(t, test.calculateStatisticsFromLastPayoutReturns, statisticsForPayout)

			nodeRepoMock.AssertNumberOfCalls(t,
				"GetAll",
				test.nodeRepoGetAllNumOfCalls,
			)
			recordRepoMock.AssertNumberOfCalls(t,
				"FindSuccessfulRecordsInsideInterval",
				test.recordRepoFindSuccessfulRecordsInsideIntervalNumOfCalls,
			)
			downtimeRepoMock.AssertNumberOfCalls(t,
				"FindDowntimesInsideInterval",
				test.downtimeRepoFindDowntimesInsideIntervalNumOfCalls,
			)
			pingRepoMock.AssertNumberOfCalls(t,
				"CalculateDowntime",
				test.pingRepoCalculateDowntimeNumOfCalls,
			)
		})
	}
}

func Test_CalculateNodeStatisticsForInterval(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		nodeID        string
		intervalStart time.Time
		intervalEnd   time.Time
		// RecordRepo.FindByNodeID
		recordRepoFindSuccessfulRecordsInsideIntervalReturns    []models.Record
		recordRepoFindSuccessfulRecordsInsideIntervalError      error
		recordRepoFindSuccessfulRecordsInsideIntervalNumOfCalls int
		// DowntimeRepo.FindByNodeID
		downtimeRepoFindDowntimesInsideIntervalReturns    []models.Downtime
		downtimeRepoFindDowntimesInsideIntervalError      error
		downtimeRepoFindDowntimesInsideIntervalNumOfCalls int
		// PingRepo.CalculateDowntime
		pingRepoCalculateDowntimeReturnDuration time.Duration
		pingRepoCalculateDowntimeError          error
		pingRepoCalculateDowntimeNumOfCalls     int
		// CalculateNodeStatisticsForInterval
		calculateNodeStatisticsForIntervalReturns *models.NodeStatsDetails
		calculateNodeStatisticsForIntervalError   error
	}{
		{
			name:   "valid statistics with multiple records and no downtime",
			nodeID: "1",
			// interval of 24 hours
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// RecordRepo.FindByNodeID
			recordRepoFindSuccessfulRecordsInsideIntervalReturns: []models.Record{
				{ID: 1, NodeId: "1", Status: "successful", Timestamp: now.Add(-20 * time.Hour)},
				{ID: 2, NodeId: "1", Status: "successful", Timestamp: now.Add(-18 * time.Hour)},
				{ID: 3, NodeId: "1", Status: "successful", Timestamp: now.Add(-17 * time.Hour)},
				{ID: 4, NodeId: "1", Status: "successful", Timestamp: now.Add(-15 * time.Hour)},
				{ID: 5, NodeId: "1", Status: "successful", Timestamp: now.Add(-12 * time.Hour)},
			},
			recordRepoFindSuccessfulRecordsInsideIntervalError:      nil,
			recordRepoFindSuccessfulRecordsInsideIntervalNumOfCalls: 1,
			// DowntimeRepo.FindByNodeID
			downtimeRepoFindDowntimesInsideIntervalReturns:    nil,
			downtimeRepoFindDowntimesInsideIntervalError:      errors.New("not found"),
			downtimeRepoFindDowntimesInsideIntervalNumOfCalls: 1,
			// PingRepo.CalculateDowntime
			pingRepoCalculateDowntimeReturnDuration: 5 * time.Second,
			pingRepoCalculateDowntimeError:          nil,
			pingRepoCalculateDowntimeNumOfCalls:     1,
			// CalculateNodeStatisticsForInterval
			calculateNodeStatisticsForIntervalReturns: &models.NodeStatsDetails{
				TotalPings:    17280, // no downtime - max number of pings
				TotalRequests: 5,
			},
			calculateNodeStatisticsForIntervalError: nil,
		},
		{
			name:          "valid statistics with no records and no downtime",
			nodeID:        "1",
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// RecordRepo.FindByNodeID
			recordRepoFindSuccessfulRecordsInsideIntervalReturns:    nil,
			recordRepoFindSuccessfulRecordsInsideIntervalError:      errors.New("not found"),
			recordRepoFindSuccessfulRecordsInsideIntervalNumOfCalls: 1,
			// DowntimeRepo.FindByNodeID
			downtimeRepoFindDowntimesInsideIntervalReturns:    nil,
			downtimeRepoFindDowntimesInsideIntervalError:      errors.New("not found"),
			downtimeRepoFindDowntimesInsideIntervalNumOfCalls: 1,
			// PingRepo.CalculateDowntime
			pingRepoCalculateDowntimeReturnDuration: 5 * time.Second,
			pingRepoCalculateDowntimeError:          nil,
			pingRepoCalculateDowntimeNumOfCalls:     1,
			// CalculateNodeStatisticsForInterval
			calculateNodeStatisticsForIntervalReturns: &models.NodeStatsDetails{
				TotalPings:    17280, // no downtime - max number of pings
				TotalRequests: 0,
			},
			calculateNodeStatisticsForIntervalError: nil,
		},
		{
			name:          "error on fetching records",
			nodeID:        "1",
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// RecordRepo.FindByNodeID
			recordRepoFindSuccessfulRecordsInsideIntervalReturns:    nil,
			recordRepoFindSuccessfulRecordsInsideIntervalError:      errors.New("db error"),
			recordRepoFindSuccessfulRecordsInsideIntervalNumOfCalls: 1,
			// CalculateNodeStatisticsForInterval
			calculateNodeStatisticsForIntervalReturns: nil,
			calculateNodeStatisticsForIntervalError:   errors.New("db error"),
		},
		{
			name:          "error on fetching pings",
			nodeID:        "1",
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// RecordRepo.FindByNodeID
			recordRepoFindSuccessfulRecordsInsideIntervalReturns:    nil,
			recordRepoFindSuccessfulRecordsInsideIntervalError:      errors.New("not found"),
			recordRepoFindSuccessfulRecordsInsideIntervalNumOfCalls: 1,
			// DowntimeRepo.FindByNodeID
			downtimeRepoFindDowntimesInsideIntervalReturns:    nil,
			downtimeRepoFindDowntimesInsideIntervalError:      errors.New("db error"),
			downtimeRepoFindDowntimesInsideIntervalNumOfCalls: 1,
			// CalculateNodeStatisticsForInterval
			calculateNodeStatisticsForIntervalReturns: nil,
			calculateNodeStatisticsForIntervalError:   errors.New("db error"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create mock controller
			recordRepoMock := mocks.RecordRepository{}
			recordRepoMock.On("FindSuccessfulRecordsInsideInterval",
				test.nodeID, test.intervalStart, test.intervalEnd,
			).Return(
				test.recordRepoFindSuccessfulRecordsInsideIntervalReturns,
				test.recordRepoFindSuccessfulRecordsInsideIntervalError,
			)
			downtimeRepoMock := mocks.DowntimeRepository{}
			downtimeRepoMock.On("FindDowntimesInsideInterval",
				test.nodeID, test.intervalStart, test.intervalEnd,
			).Return(
				test.downtimeRepoFindDowntimesInsideIntervalReturns,
				test.downtimeRepoFindDowntimesInsideIntervalError,
			)
			pingRepoMock := mocks.PingRepository{}
			pingRepoMock.On("CalculateDowntime",
				test.nodeID, test.intervalEnd,
			).Return(
				time.Now(),
				test.pingRepoCalculateDowntimeReturnDuration,
				test.pingRepoCalculateDowntimeError,
			)
			repos := repositories.Repos{
				PingRepo:     &pingRepoMock,
				RecordRepo:   &recordRepoMock,
				DowntimeRepo: &downtimeRepoMock,
			}

			statisticsForPayout, err := CalculateNodeStatisticsForInterval(repos, test.nodeID, test.intervalStart, test.intervalEnd)

			assert.Equal(t, test.calculateNodeStatisticsForIntervalError, err)
			assert.Equal(t, test.calculateNodeStatisticsForIntervalReturns, statisticsForPayout)

			recordRepoMock.AssertNumberOfCalls(t,
				"FindSuccessfulRecordsInsideInterval",
				test.recordRepoFindSuccessfulRecordsInsideIntervalNumOfCalls,
			)
			downtimeRepoMock.AssertNumberOfCalls(t,
				"FindDowntimesInsideInterval",
				test.downtimeRepoFindDowntimesInsideIntervalNumOfCalls,
			)
			pingRepoMock.AssertNumberOfCalls(t,
				"CalculateDowntime",
				test.pingRepoCalculateDowntimeNumOfCalls,
			)
		})
	}
}

func Test_CalculateStatisticsForInterval(t *testing.T) {
	now := time.Now()

	testNode := models.Node{
		ID: "1",
		PayoutAddress: "0xpayout-address-1",
	}

	tests := []struct {
		name          string
		intervalStart time.Time
		intervalEnd   time.Time
		// NodeRepo.GetAll
		nodeRepoGetAllReturns    *[]models.Node
		nodeRepoGetAllError      error
		nodeRepoGetAllNumOfCalls int
		// RecordRepo.FindByNodeID
		recordRepoFindSuccessfulRecordsInsideIntervalReturns    []models.Record
		recordRepoFindSuccessfulRecordsInsideIntervalError      error
		recordRepoFindSuccessfulRecordsInsideIntervalNumOfCalls int
		// DowntimeRepo.FindByNodeID
		downtimeRepoFindDowntimesInsideIntervalReturns    []models.Downtime
		downtimeRepoFindDowntimesInsideIntervalError      error
		downtimeRepoFindDowntimesInsideIntervalNumOfCalls int
		// PingRepo.CalculateDowntime
		pingRepoCalculateDowntimeReturnDuration time.Duration
		pingRepoCalculateDowntimeError          error
		pingRepoCalculateDowntimeNumOfCalls     int
		// CalculateNodeStatisticsForInterval
		calculateStatisticsForIntervalReturns map[string]models.NodeStatsDetails
		calculateStatisticsForIntervalError   error
	}{
		{
			name:   "valid statistics with multiple records and no downtime",
			// interval of 24 hours
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// NodeRepo.GetAll
			nodeRepoGetAllReturns: &[]models.Node{
				testNode,
			},
			nodeRepoGetAllError:      nil,
			nodeRepoGetAllNumOfCalls: 1,
			// RecordRepo.FindByNodeID
			recordRepoFindSuccessfulRecordsInsideIntervalReturns: []models.Record{
				{ID: 1, NodeId: testNode.ID, Status: "successful", Timestamp: now.Add(-20 * time.Hour)},
				{ID: 2, NodeId: testNode.ID, Status: "successful", Timestamp: now.Add(-18 * time.Hour)},
				{ID: 3, NodeId: testNode.ID, Status: "successful", Timestamp: now.Add(-17 * time.Hour)},
				{ID: 4, NodeId: testNode.ID, Status: "successful", Timestamp: now.Add(-15 * time.Hour)},
				{ID: 5, NodeId: testNode.ID, Status: "successful", Timestamp: now.Add(-12 * time.Hour)},
			},
			recordRepoFindSuccessfulRecordsInsideIntervalError:      nil,
			recordRepoFindSuccessfulRecordsInsideIntervalNumOfCalls: 1,
			// DowntimeRepo.FindByNodeID
			downtimeRepoFindDowntimesInsideIntervalReturns:    nil,
			downtimeRepoFindDowntimesInsideIntervalError:      errors.New("not found"),
			downtimeRepoFindDowntimesInsideIntervalNumOfCalls: 1,
			// PingRepo.CalculateDowntime
			pingRepoCalculateDowntimeReturnDuration: 5 * time.Second,
			pingRepoCalculateDowntimeError:          nil,
			pingRepoCalculateDowntimeNumOfCalls:     1,
			// CalculateNodeStatisticsForInterval
			calculateStatisticsForIntervalReturns: map[string]models.NodeStatsDetails{
				testNode.PayoutAddress: {
					TotalPings:    17280, // no downtime - max number of pings
					TotalRequests: 5,
				},
			},
			calculateStatisticsForIntervalError: nil,
		},
		{
			name:   "get all nodes fails",
			// interval of 24 hours
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// NodeRepo.GetAll
			nodeRepoGetAllReturns:    nil,
			nodeRepoGetAllError:      errors.New("not found"),
			nodeRepoGetAllNumOfCalls: 1,
			// CalculateNodeStatisticsForInterval
			calculateStatisticsForIntervalReturns: nil,
			calculateStatisticsForIntervalError:   errors.New("not found"),
		},
		{
			name:   "calculating statistics for node fails",
			// interval of 24 hours
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// NodeRepo.GetAll
			nodeRepoGetAllReturns: &[]models.Node{
				testNode,
			},
			nodeRepoGetAllError:      nil,
			nodeRepoGetAllNumOfCalls: 1,
			// RecordRepo.FindByNodeID
			recordRepoFindSuccessfulRecordsInsideIntervalReturns:    nil,
			recordRepoFindSuccessfulRecordsInsideIntervalError:      errors.New("db error"),
			recordRepoFindSuccessfulRecordsInsideIntervalNumOfCalls: 1,
			// CalculateNodeStatisticsForInterval
			calculateStatisticsForIntervalReturns: nil,
			calculateStatisticsForIntervalError:   errors.New("db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create mock controller
			nodeRepoMock := mocks.NodeRepository{}
			nodeRepoMock.On("GetAll").Return(
				test.nodeRepoGetAllReturns, test.nodeRepoGetAllError,
			)
			recordRepoMock := mocks.RecordRepository{}
			recordRepoMock.On("FindSuccessfulRecordsInsideInterval",
				testNode.ID, test.intervalStart, test.intervalEnd,
			).Return(
				test.recordRepoFindSuccessfulRecordsInsideIntervalReturns,
				test.recordRepoFindSuccessfulRecordsInsideIntervalError,
			)
			downtimeRepoMock := mocks.DowntimeRepository{}
			downtimeRepoMock.On("FindDowntimesInsideInterval",
				testNode.ID, test.intervalStart, test.intervalEnd,
			).Return(
				test.downtimeRepoFindDowntimesInsideIntervalReturns,
				test.downtimeRepoFindDowntimesInsideIntervalError,
			)
			pingRepoMock := mocks.PingRepository{}
			pingRepoMock.On("CalculateDowntime",
				testNode.ID, test.intervalEnd,
			).Return(
				time.Now(),
				test.pingRepoCalculateDowntimeReturnDuration,
				test.pingRepoCalculateDowntimeError,
			)
			repos := repositories.Repos{
				PingRepo:     &pingRepoMock,
				RecordRepo:   &recordRepoMock,
				DowntimeRepo: &downtimeRepoMock,
				NodeRepo:     &nodeRepoMock,
			}

			statisticsForPayout, err := CalculateStatisticsForInterval(repos, test.intervalStart, test.intervalEnd)

			assert.Equal(t, test.calculateStatisticsForIntervalError, err)
			assert.Equal(t, test.calculateStatisticsForIntervalReturns, statisticsForPayout)

			nodeRepoMock.AssertNumberOfCalls(t,
				"GetAll",
				test.nodeRepoGetAllNumOfCalls,
			)
			recordRepoMock.AssertNumberOfCalls(t,
				"FindSuccessfulRecordsInsideInterval",
				test.recordRepoFindSuccessfulRecordsInsideIntervalNumOfCalls,
			)
			downtimeRepoMock.AssertNumberOfCalls(t,
				"FindDowntimesInsideInterval",
				test.downtimeRepoFindDowntimesInsideIntervalNumOfCalls,
			)
			pingRepoMock.AssertNumberOfCalls(t,
				"CalculateDowntime",
				test.pingRepoCalculateDowntimeNumOfCalls,
			)
		})
	}
}
