package stats

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	mocks "github.com/NodeFactoryIo/vedran/mocks/repositories"
	"github.com/magiconair/properties/assert"
	"testing"
	"time"
)

func TestCalculateStatisticsForPayout(t *testing.T) {
	tests := []struct {
		name string
		// NodeRepo.GetAll
		nodeRepoGetAllReturns []models.Node
		nodeRepoGetAllError error
		// RecordRepo.FindByNodeID
		recordRepoFindByNodeIDAndIntervalReturns []models.Record
		recordRepoFindByNodeIDAndIntervalError error
		// DowntimeRepo.FindByNodeID
		downtimeRepoFindByNodeIDAndIntervalReturns []models.Downtime
		downtimeRepoFindByNodeIDAndIntervalError error
		// PingRepo.CalculateDowntime
		pingRepoCalculateDowntimeReturnDuration time.Duration
		pingRepoCalculateDowntimeReturnError error
		//
		calculateStatisticsForPayoutError error
		calculateStatisticsForPayoutReturns map[string]models.NodePaymentDetails
	}{
		{
			name: "",
			nodeRepoGetAllReturns: nil,
			nodeRepoGetAllError: nil,
			recordRepoFindByNodeIDAndIntervalReturns: nil,
			recordRepoFindByNodeIDAndIntervalError: nil,
			downtimeRepoFindByNodeIDAndIntervalReturns: nil,
			downtimeRepoFindByNodeIDAndIntervalError: nil,
			pingRepoCalculateDowntimeReturnDuration: nil,
			pingRepoCalculateDowntimeReturnError: nil,
		},
		{
			name: "",
			nodeRepoGetAllReturns: nil,
			nodeRepoGetAllError: nil,
			recordRepoFindByNodeIDAndIntervalReturns: nil,
			recordRepoFindByNodeIDAndIntervalError: nil,
			downtimeRepoFindByNodeIDAndIntervalReturns: nil,
			downtimeRepoFindByNodeIDAndIntervalError: nil,
			pingRepoCalculateDowntimeReturnDuration: nil,
			pingRepoCalculateDowntimeReturnError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create mock controller
			nodeRepoMock := mocks.NodeRepository{}
			recordRepoMock := mocks.RecordRepository{}
			metricsRepoMock := mocks.MetricsRepository{}
			pingRepoMock := mocks.PingRepository{}
			downtimeRepoMock := mocks.DowntimeRepository{}
			paymentRepoMock := mocks.PaymentRepository{}
			repos := repositories.Repos{
				NodeRepo:     &nodeRepoMock,
				PingRepo:     &pingRepoMock,
				MetricsRepo:  &metricsRepoMock,
				RecordRepo:   &recordRepoMock,
				DowntimeRepo: &downtimeRepoMock,
				PaymentRepo:  &paymentRepoMock,
			}

			statisticsForPayout, err := CalculateStatisticsForPayout(repos)
			assert.Equal(t, err, test.calculateStatisticsForPayoutError)
			assert.Equal(t, statisticsForPayout, test.calculateStatisticsForPayoutReturns)
		})
	}
}
