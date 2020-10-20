package active

import (
	"errors"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	mocks "github.com/NodeFactoryIo/vedran/mocks/repositories"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCheckIfNodeActive(t *testing.T) {
	tests := []struct {
		name string
		node models.Node
		nodePing *models.Ping
		nodePingError interface{}
		nodeMetrics *models.Metrics
		nodeMetricsError interface{}
		latestMetrics *models.LatestBlockMetrics
		latestMetricsError interface{}
		expectedResult bool
		expectedError interface{}
	}{
		{
			name: "active node",
			node: models.Node{ID: "1"},
			nodePing: &models.Ping{
				NodeId:    "1",
				Timestamp: time.Now(),
			},
			nodePingError: nil,
			nodeMetrics: &models.Metrics{
				NodeId:                "1",
				PeerCount:             0,
				BestBlockHeight:       1000,
				FinalizedBlockHeight:  995,
				ReadyTransactionCount: 0,
			},
			nodeMetricsError: nil,
			latestMetrics: &models.LatestBlockMetrics{
				BestBlockHeight:      1001,
				FinalizedBlockHeight: 998,
			},
			latestMetricsError: nil,
			expectedResult: true,
			expectedError: nil,
		},
		{
			name: "not active node::ping old",
			node: models.Node{ID: "1"},
			nodePing: &models.Ping{
				NodeId:    "1",
				Timestamp: time.Unix(10, 10),
			},
			nodePingError: nil,
			nodeMetrics: &models.Metrics{
				NodeId:                "1",
				PeerCount:             0,
				BestBlockHeight:       1000,
				FinalizedBlockHeight:  995,
				ReadyTransactionCount: 0,
			},
			nodeMetricsError: nil,
			latestMetrics: &models.LatestBlockMetrics{
				BestBlockHeight:      0,
				FinalizedBlockHeight: 0,
			},
			latestMetricsError: nil,
			expectedResult: false,
			expectedError: nil,
		},
		{
			name: "not active node::bad metrics",
			node: models.Node{ID: "1"},
			nodePing: &models.Ping{
				NodeId:    "1",
				Timestamp: time.Now(),
			},
			nodePingError: nil,
			nodeMetrics: &models.Metrics{
				NodeId:                "1",
				PeerCount:             0,
				BestBlockHeight:       1000,
				FinalizedBlockHeight:  995,
				ReadyTransactionCount: 0,
			},
			nodeMetricsError: nil,
			latestMetrics: &models.LatestBlockMetrics{
				BestBlockHeight:      1200,
				FinalizedBlockHeight: 1192,
			},
			latestMetricsError: nil,
			expectedResult: false,
			expectedError: nil,
		},
		{
			name: "not active node::bad metric finalized block",
			node: models.Node{ID: "1"},
			nodePing: &models.Ping{
				NodeId:    "1",
				Timestamp: time.Now(),
			},
			nodePingError: nil,
			nodeMetrics: &models.Metrics{
				NodeId:                "1",
				PeerCount:             0,
				BestBlockHeight:       1000,
				FinalizedBlockHeight:  994,
				ReadyTransactionCount: 0,
			},
			nodeMetricsError: nil,
			latestMetrics: &models.LatestBlockMetrics{
				BestBlockHeight:      1009,
				FinalizedBlockHeight: 1005,
			},
			latestMetricsError: nil,
			expectedResult: false,
			expectedError: nil,
		},
		{
			name: "not active node::bad metric best block",
			node: models.Node{ID: "1"},
			nodePing: &models.Ping{
				NodeId:    "1",
				Timestamp: time.Now(),
			},
			nodePingError: nil,
			nodeMetrics: &models.Metrics{
				NodeId:                "1",
				PeerCount:             0,
				BestBlockHeight:       1000,
				FinalizedBlockHeight:  995,
				ReadyTransactionCount: 0,
			},
			nodeMetricsError: nil,
			latestMetrics: &models.LatestBlockMetrics{
				BestBlockHeight:      1011,
				FinalizedBlockHeight: 1000,
			},
			latestMetricsError: nil,
			expectedResult: false,
			expectedError: nil,
		},
		{
			name: "ping repo fails",
			node: models.Node{ID: "1"},
			nodePing: &models.Ping{
				NodeId:    "1",
				Timestamp: time.Now(),
			},
			nodePingError: errors.New("ping-error"),
			nodeMetrics: &models.Metrics{
				NodeId:                "1",
				PeerCount:             0,
				BestBlockHeight:       1000,
				FinalizedBlockHeight:  995,
				ReadyTransactionCount: 0,
			},
			nodeMetricsError: nil,
			latestMetrics: &models.LatestBlockMetrics{
				BestBlockHeight:      1001,
				FinalizedBlockHeight: 998,
			},
			latestMetricsError: nil,
			expectedResult: false,
			expectedError: errors.New("ping-error"),
		},
		{
			name: "metrics repo fails on node metrics",
			node: models.Node{ID: "1"},
			nodePing: &models.Ping{
				NodeId:    "1",
				Timestamp: time.Now(),
			},
			nodePingError: nil,
			nodeMetrics: &models.Metrics{
				NodeId:                "1",
				PeerCount:             0,
				BestBlockHeight:       1000,
				FinalizedBlockHeight:  995,
				ReadyTransactionCount: 0,
			},
			nodeMetricsError: errors.New("metrics-error"),
			latestMetrics: &models.LatestBlockMetrics{
				BestBlockHeight:      1001,
				FinalizedBlockHeight: 998,
			},
			latestMetricsError: nil,
			expectedResult: false,
			expectedError: errors.New("metrics-error"),
		},
		{
			name: "metrics repo fails on latest metrics",
			node: models.Node{ID: "1"},
			nodePing: &models.Ping{
				NodeId:    "1",
				Timestamp: time.Now(),
			},
			nodePingError: nil,
			nodeMetrics: &models.Metrics{
				NodeId:                "1",
				PeerCount:             0,
				BestBlockHeight:       1000,
				FinalizedBlockHeight:  995,
				ReadyTransactionCount: 0,
			},
			nodeMetricsError: nil,
			latestMetrics: &models.LatestBlockMetrics{
				BestBlockHeight:      1001,
				FinalizedBlockHeight: 998,
			},
			latestMetricsError: errors.New("metrics-error"),
			expectedResult: false,
			expectedError: errors.New("metrics-error"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create mock controller
			nodeRepoMock := mocks.NodeRepository{}
			pingRepoMock := mocks.PingRepository{}
			pingRepoMock.On("FindByNodeID", test.node.ID).Return(test.nodePing, test.nodePingError)
			metricsRepoMock := mocks.MetricsRepository{}
			metricsRepoMock.On("FindByID", test.node.ID).Return(test.nodeMetrics, test.nodeMetricsError)
			metricsRepoMock.On("GetLatestBlockMetrics").Return(test.latestMetrics, test.latestMetricsError)
			recordRepoMock := mocks.RecordRepository{}

			result, err := CheckIfNodeActive(test.node, &repositories.Repos{
				NodeRepo:    &nodeRepoMock,
				PingRepo:    &pingRepoMock,
				MetricsRepo: &metricsRepoMock,
				RecordRepo:  &recordRepoMock,
			})

			assert.Equal(t, result, test.expectedResult)
			assert.Equal(t, err, test.expectedError)
		})
	}
}
