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

func Test_CalculateTotalPingsForNode(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		nodeID        string
		intervalStart time.Time
		intervalEnd   time.Time
		// DowntimeRepo.FindByNodeID
		downtimeRepoFindDowntimesInsideIntervalReturns    []models.Downtime
		downtimeRepoFindDowntimesInsideIntervalError      error
		downtimeRepoFindDowntimesInsideIntervalNumOfCalls int
		// PingRepo.CalculateDowntime
		pingRepoCalculateDowntimeReturnDuration time.Duration
		pingRepoCalculateDowntimeError          error
		pingRepoCalculateDowntimeNumOfCalls     int
		//
		calculateTotalPingsForNodeError   error
		calculateTotalPingsForNodeReturns float64
	}{
		{
			name:   "no downtimes",
			nodeID: "1",
			// interval of 24 hours
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// DowntimeRepo.FindByNodeID
			downtimeRepoFindDowntimesInsideIntervalReturns: nil,
			downtimeRepoFindDowntimesInsideIntervalError:      errors.New("not found"),
			downtimeRepoFindDowntimesInsideIntervalNumOfCalls: 1,
			// PingRepo.CalculateDowntime
			pingRepoCalculateDowntimeReturnDuration: 5 * time.Second,
			pingRepoCalculateDowntimeError:          nil,
			pingRepoCalculateDowntimeNumOfCalls:     1,
			// [total_interval - total_downtime]   / ping_interval = num_of_pings
			// [24h (86400s) - 0min (0s)]          / 10            = 8640
			calculateTotalPingsForNodeReturns: float64(86400 / PingIntervalInSeconds),
			calculateTotalPingsForNodeError:   nil,
		},
		{
			name:   "multiple downtimes inside interval",
			nodeID: "1",
			// interval of 24 hours
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// DowntimeRepo.FindByNodeID
			downtimeRepoFindDowntimesInsideIntervalReturns: []models.Downtime{
				{ // downtime 1h
					ID:     1,
					NodeId: "1",
					Start:  now.Add(-11 * time.Hour),
					End:    now.Add(-10 * time.Hour),
				},
				{ // downtime 20min
					ID:     2,
					NodeId: "1",
					Start:  now.Add(-120 * time.Minute),
					End:    now.Add(-100 * time.Minute),
				},
				{ // downtime 20s
					ID: 3,
					NodeId: "1",
					Start: now.Add(-120 * time.Second),
					End: now.Add(-100 * time.Second),
				},
			},
			downtimeRepoFindDowntimesInsideIntervalError:      nil,
			downtimeRepoFindDowntimesInsideIntervalNumOfCalls: 1,
			// PingRepo.CalculateDowntime
			pingRepoCalculateDowntimeReturnDuration: 1 * time.Second,
			pingRepoCalculateDowntimeError:          nil,
			pingRepoCalculateDowntimeNumOfCalls:     1,
			// [total_interval - total_downtime]   / ping_interval = num_of_pings
			// [24h (86400s) - 1h20min20s (4820s)] / 10            = 8158
			calculateTotalPingsForNodeReturns: float64(81580 / PingIntervalInSeconds),
			calculateTotalPingsForNodeError:   nil,
		},
		{
			name:   "downtime started before interval",
			nodeID: "1",
			// interval of 24 hours
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// DowntimeRepo.FindByNodeID
			downtimeRepoFindDowntimesInsideIntervalReturns: []models.Downtime{
				{ // downtime 1h30min (30min effective), started before calculated interval
					ID:     1,
					NodeId: "1",
					Start:  now.Add(-1500 * time.Minute), // -24h60min
					End:    now.Add(-1410 * time.Minute),   // -23h30min
				},
			},
			downtimeRepoFindDowntimesInsideIntervalError:      nil,
			downtimeRepoFindDowntimesInsideIntervalNumOfCalls: 1,
			// PingRepo.CalculateDowntime
			pingRepoCalculateDowntimeReturnDuration: 1 * time.Second,
			pingRepoCalculateDowntimeError:          nil,
			pingRepoCalculateDowntimeNumOfCalls:     1,
			// [total_interval - total_downtime]   / ping_interval = num_of_pings
			// [24h (86400s) - 30min (1800s)]      / 10            = 8460
			calculateTotalPingsForNodeReturns: float64(84600 / PingIntervalInSeconds),
			calculateTotalPingsForNodeError:   nil,
		},
		{
			name:   "downtime still active",
			nodeID: "1",
			// interval of 24 hours
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// DowntimeRepo.FindByNodeID
			downtimeRepoFindDowntimesInsideIntervalReturns: []models.Downtime{},
			downtimeRepoFindDowntimesInsideIntervalError:      nil,
			downtimeRepoFindDowntimesInsideIntervalNumOfCalls: 1,
			// PingRepo.CalculateDowntime
			pingRepoCalculateDowntimeReturnDuration: 30 * time.Minute,
			pingRepoCalculateDowntimeError:          nil,
			pingRepoCalculateDowntimeNumOfCalls:     1,
			// [total_interval - total_downtime]   / ping_interval = num_of_pings
			// [24h (86400s) - 30min (1800s)]      / 10            = 8460
			calculateTotalPingsForNodeReturns: float64(84600 / PingIntervalInSeconds),
			calculateTotalPingsForNodeError:   nil,
		},
		{
			name:   "mixed multiple downtimes",
			nodeID: "1",
			// interval of 24 hours
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// DowntimeRepo.FindByNodeID
			downtimeRepoFindDowntimesInsideIntervalReturns: []models.Downtime{
				{ // downtime 1h20min (20min effective), started before calculated interval
					ID:     1,
					NodeId: "1",
					Start:  now.Add(-1500 * time.Minute),   // -24h60min
					End:    now.Add(-1420 * time.Minute),   // -23h40min
				},
				{ // downtime 20min
					ID:     2,
					NodeId: "1",
					Start:  now.Add(-220 * time.Minute),
					End:    now.Add(-200 * time.Minute),
				},
			},
			downtimeRepoFindDowntimesInsideIntervalError:      nil,
			downtimeRepoFindDowntimesInsideIntervalNumOfCalls: 1,
			// PingRepo.CalculateDowntime
			pingRepoCalculateDowntimeReturnDuration: 20 * time.Minute,
			pingRepoCalculateDowntimeError:          nil,
			pingRepoCalculateDowntimeNumOfCalls:     1,
			// [total_interval - total_downtime]   / ping_interval = num_of_pings
			// [24h (86400s) - 3x20min (3600s)]    / 10            = 8280
			calculateTotalPingsForNodeReturns: float64(82800 / PingIntervalInSeconds),
			calculateTotalPingsForNodeError:   nil,
		},
		{
			name:   "error on fetching downtime",
			nodeID: "1",
			// interval of 24 hours
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// DowntimeRepo.FindByNodeID
			downtimeRepoFindDowntimesInsideIntervalReturns: nil,
			downtimeRepoFindDowntimesInsideIntervalError:      errors.New("db error"),
			downtimeRepoFindDowntimesInsideIntervalNumOfCalls: 1,
			// CalculateTotalPingsForNode
			calculateTotalPingsForNodeReturns: float64(0),
			calculateTotalPingsForNodeError:   errors.New("db error"),
		},
		{
			name:   "error on fetching downtime",
			nodeID: "1",
			// interval of 24 hours
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// DowntimeRepo.FindByNodeID
			downtimeRepoFindDowntimesInsideIntervalReturns: []models.Downtime{},
			downtimeRepoFindDowntimesInsideIntervalError:      nil,
			downtimeRepoFindDowntimesInsideIntervalNumOfCalls: 1,
			// PingRepo.CalculateDowntime
			pingRepoCalculateDowntimeReturnDuration: 0 * time.Second,
			pingRepoCalculateDowntimeError:          errors.New("db error"),
			pingRepoCalculateDowntimeNumOfCalls:     1,
			// CalculateTotalPingsForNode
			calculateTotalPingsForNodeReturns: float64(0),
			calculateTotalPingsForNodeError:   errors.New("db error"),
		},
		{
			name:   "downtime bigger than interval",
			nodeID: "1",
			// interval of 24 hours
			intervalStart: now.Add(-24 * time.Hour),
			intervalEnd:   now,
			// DowntimeRepo.FindByNodeID
			downtimeRepoFindDowntimesInsideIntervalReturns: []models.Downtime{},
			downtimeRepoFindDowntimesInsideIntervalError:      nil,
			downtimeRepoFindDowntimesInsideIntervalNumOfCalls: 1,
			// PingRepo.CalculateDowntime
			pingRepoCalculateDowntimeReturnDuration: 26 * time.Hour,
			pingRepoCalculateDowntimeError:          nil,
			pingRepoCalculateDowntimeNumOfCalls:     1,
			// CalculateTotalPingsForNode
			calculateTotalPingsForNodeReturns: float64(0),
			calculateTotalPingsForNodeError:   nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

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
				DowntimeRepo: &downtimeRepoMock,
			}

			totalPings, err := CalculateTotalPingsForNode(repos, test.nodeID, test.intervalStart, test.intervalEnd)

			assert.Equal(t, err, test.calculateTotalPingsForNodeError)
			assert.Equal(t, test.calculateTotalPingsForNodeReturns, totalPings)

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
