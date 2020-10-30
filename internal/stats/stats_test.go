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

func TestCalculateStatisticsForNode(t *testing.T) {
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
		// CalculateStatisticsForNode
		calculateStatisticsForNodeReturns *models.NodePaymentDetails
		calculateStatisticsForNodeError   error
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
			// CalculateStatisticsForNode
			calculateStatisticsForNodeReturns: &models.NodePaymentDetails{
				TotalPings:    8640, // no downtime - max number of pings
				TotalRequests: 5,
			},
			calculateStatisticsForNodeError: nil,
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
			// CalculateStatisticsForNode
			calculateStatisticsForNodeReturns: &models.NodePaymentDetails{
				TotalPings:    8640, // no downtime - max number of pings
				TotalRequests: 0,
			},
			calculateStatisticsForNodeError: nil,
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
			// CalculateStatisticsForNode
			calculateStatisticsForNodeReturns: nil,
			calculateStatisticsForNodeError:   errors.New("db error"),
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
			// CalculateStatisticsForNode
			calculateStatisticsForNodeReturns: nil,
			calculateStatisticsForNodeError:   errors.New("db error"),
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

			statisticsForPayout, err := CalculateStatisticsForNode(repos, test.nodeID, test.intervalStart, test.intervalEnd)

			assert.Equal(t, test.calculateStatisticsForNodeError, err)
			assert.Equal(t, test.calculateStatisticsForNodeReturns, statisticsForPayout)

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

func TestCalculateStatisticsForPayout(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		nodeID        string
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
		// CalculateStatisticsForNode
		calculateStatisticsForPayoutReturns map[string]models.NodePaymentDetails
		calculateStatisticsForPayoutError   error
	}{
		{
			name:   "valid statistics with multiple records and no downtime",
			nodeID: "1",
			// interval of 24 hours
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// NodeRepo.GetAll
			nodeRepoGetAllReturns: &[]models.Node{
				{
					ID: "1",
				},
			},
			nodeRepoGetAllError:      nil,
			nodeRepoGetAllNumOfCalls: 1,
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
			// CalculateStatisticsForNode
			calculateStatisticsForPayoutReturns: map[string]models.NodePaymentDetails{
				"1": {
					TotalPings:    8640, // no downtime - max number of pings
					TotalRequests: 5,
				},
			},
			calculateStatisticsForPayoutError: nil,
		},
		{
			name:   "get all nodes fails",
			nodeID: "1",
			// interval of 24 hours
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// NodeRepo.GetAll
			nodeRepoGetAllReturns: nil,
			nodeRepoGetAllError:      errors.New("not found"),
			nodeRepoGetAllNumOfCalls: 1,
			// CalculateStatisticsForNode
			calculateStatisticsForPayoutReturns: nil,
			calculateStatisticsForPayoutError: errors.New("not found"),
		},
		{
			name:   "calculating statistics for node fails",
			nodeID: "1",
			// interval of 24 hours
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// NodeRepo.GetAll
			nodeRepoGetAllReturns: &[]models.Node{
				{
					ID: "1",
				},
			},
			nodeRepoGetAllError:      nil,
			nodeRepoGetAllNumOfCalls: 1,
			// RecordRepo.FindByNodeID
			recordRepoFindSuccessfulRecordsInsideIntervalReturns:    nil,
			recordRepoFindSuccessfulRecordsInsideIntervalError:      errors.New("db error"),
			recordRepoFindSuccessfulRecordsInsideIntervalNumOfCalls: 1,
			// CalculateStatisticsForNode
			calculateStatisticsForPayoutReturns: nil,
			calculateStatisticsForPayoutError:   errors.New("db error"),
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
				NodeRepo:     &nodeRepoMock,
			}

			statisticsForPayout, err := CalculateStatisticsForPayout(repos, test.intervalStart, test.intervalEnd)

			assert.Equal(t, test.calculateStatisticsForPayoutError, err)
			assert.Equal(t, test.calculateStatisticsForPayoutReturns, statisticsForPayout)

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
