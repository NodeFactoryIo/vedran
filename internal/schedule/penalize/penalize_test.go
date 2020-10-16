package penalize

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	repoMocks "github.com/NodeFactoryIo/vedran/mocks/repositories"
	"testing"
	"time"
)

func TestScheduleCheckForPenalizedNode(t *testing.T) {
	tests := []struct {
		name                          string
		nodeID                        string
		node                          models.Node
		saveNode                      []*models.Node
		saveNodeNumberOfCalls         int
		addToActiveNode               []models.Node
		addToActiveNodesNumberOfCalls int
		nodePing                      []*models.Ping
		nodeMetrics                   []*models.Metrics
		latestMetrics                 []*models.LatestBlockMetrics
	}{
		{
			name:   "penalized node becomes active on first check",
			nodeID: "1",
			node: models.Node{
				ID:       "1",
				Cooldown: 1,
			},
			saveNode: nil,
			saveNodeNumberOfCalls: 0,
			addToActiveNode: []models.Node{
				{
					ID:       "1",
					Cooldown: 1,
				},
			},
			addToActiveNodesNumberOfCalls: 1,
			nodePing: []*models.Ping{
				{
					NodeId:    "1",
					Timestamp: time.Now(),
				},
			},
			nodeMetrics: []*models.Metrics{
				{
					NodeId:               "",
					BestBlockHeight:      1000,
					FinalizedBlockHeight: 995,
				},
			},
			latestMetrics: []*models.LatestBlockMetrics{
				{
					BestBlockHeight:      1001,
					FinalizedBlockHeight: 998,
				},
			},
		},
		{
			name:   "penalized node schedule one check",
			nodeID: "1",
			node: models.Node{
				ID:       "1",
				Cooldown: 1,
			},
			saveNode: []*models.Node{
				{
					ID:       "1",
					Cooldown: 2,
				},
			},
			saveNodeNumberOfCalls: 1,
			addToActiveNode: []models.Node{
				{
					ID:       "1",
					Cooldown: 2,
				},
			},
			addToActiveNodesNumberOfCalls: 1,
			nodePing: []*models.Ping{
				{
					NodeId:    "1",
					Timestamp: time.Now(),
				},
			},
			nodeMetrics: []*models.Metrics{
				{
					NodeId:               "1",
					BestBlockHeight:      900,
					FinalizedBlockHeight: 898,
				},
				{
					NodeId:               "",
					BestBlockHeight:      1000,
					FinalizedBlockHeight: 995,
				},
			},
			latestMetrics: []*models.LatestBlockMetrics{
				{
					BestBlockHeight:      1001,
					FinalizedBlockHeight: 998,
				}, {
					BestBlockHeight:      1001,
					FinalizedBlockHeight: 998,
				},
			},
		},
		{
			name:   "penalized node schedule multiple checks",
			nodeID: "1",
			node: models.Node{
				ID:       "1",
				Cooldown: 1,
			},
			saveNode: []*models.Node{
				{
					ID:       "1",
					Cooldown: 2,
				},
				{
					ID:       "1",
					Cooldown: 4,
				},
				{
					ID:       "1",
					Cooldown: 8,
				},
			},
			saveNodeNumberOfCalls: 3,
			addToActiveNode: []models.Node{
				{
					ID:       "1",
					Cooldown: 8,
				},
			},
			addToActiveNodesNumberOfCalls: 1,
			nodePing: []*models.Ping{
				{
					NodeId:    "1",
					Timestamp: time.Now(),
				},
			},
			nodeMetrics: []*models.Metrics{
				{
					NodeId:               "1",
					BestBlockHeight:      900,
					FinalizedBlockHeight: 898,
				},
				{
					NodeId:               "1",
					BestBlockHeight:      900,
					FinalizedBlockHeight: 898,
				},
				{
					NodeId:               "1",
					BestBlockHeight:      900,
					FinalizedBlockHeight: 898,
				},
				{
					NodeId:               "",
					BestBlockHeight:      1000,
					FinalizedBlockHeight: 995,
				},
			},
			latestMetrics: []*models.LatestBlockMetrics{
				{
					BestBlockHeight:      1001,
					FinalizedBlockHeight: 998,
				},
			},
		},
		{
			name:   "penalized node hits max cooldown",
			nodeID: "1",
			node: models.Node{
				ID:       "1",
				Cooldown: 510,
			},
			saveNode: []*models.Node{
				{
					ID:       "1",
					Cooldown: 1020, // 1020 is MaxCooldownForPenalizedNode in minutes (17 hours)
				},
			},
			saveNodeNumberOfCalls: 1,
			addToActiveNode: nil,
			addToActiveNodesNumberOfCalls: 0,
			nodePing: []*models.Ping{
				{
					NodeId:    "1",
					Timestamp: time.Now(),
				},
			},
			nodeMetrics: []*models.Metrics{
				{
					NodeId:               "1",
					BestBlockHeight:      900,
					FinalizedBlockHeight: 898,
				},
			},
			latestMetrics: []*models.LatestBlockMetrics{
				{
					BestBlockHeight:      1001,
					FinalizedBlockHeight: 998,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			nodeRepoMock := repoMocks.NodeRepository{}

			// is mocked function called in test
			if test.addToActiveNode != nil {
				if len(test.addToActiveNode) == 1 { // same return value
					nodeRepoMock.On("AddNodeToActive", test.addToActiveNode[0]).Return()
				} else { // multiple return values
					for _, n := range test.addToActiveNode {
						nodeRepoMock.On("AddNodeToActive", n).Return().Once()
					}
				}
			}

			// is mocked function called in test
			if test.saveNode != nil {
				if len(test.saveNode) == 1 { // same return value
					nodeRepoMock.On("Save", test.saveNode[0]).Return(nil)
				} else { // multiple return values
					for _, nn := range test.saveNode {
						nodeRepoMock.On("Save", nn).Return(nil).Once()
					}
				}
			}

			pingRepoMock := repoMocks.PingRepository{}
			if len(test.nodePing) == 1 { // same return value
				pingRepoMock.On("FindByNodeID", test.nodeID).Return(test.nodePing[0], nil)
			} else { // multiple return values
				for _, ping := range test.nodePing {
					pingRepoMock.On("FindByNodeID", test.nodeID).Return(ping, nil).Once()
				}
			}

			metricsRepoMock := repoMocks.MetricsRepository{}

			if len(test.nodeMetrics) == 1 { // same return value
				metricsRepoMock.On("FindByID", test.nodeID).Return(test.nodeMetrics[0], nil)
			} else { // multiple return values
				for _, metric := range test.nodeMetrics {
					metricsRepoMock.On("FindByID", test.nodeID).Return(metric, nil).Once()
				}
			}

			if len(test.latestMetrics) == 1 { // same return value
				metricsRepoMock.On("GetLatestBlockMetrics").Return(test.latestMetrics[0], nil)
			} else {
				for _, latestMetric := range test.latestMetrics { // multiple return values
					metricsRepoMock.On("GetLatestBlockMetrics").Return(latestMetric, nil).Once()
				}
			}

			recordRepoMock := repoMocks.RecordRepository{}

			afterFunc = func(d time.Duration, f func()) *time.Timer {
				f()
				return nil
			}

			ScheduleCheckForPenalizedNode(test.node, repositories.Repos{
				NodeRepo:    &nodeRepoMock,
				PingRepo:    &pingRepoMock,
				MetricsRepo: &metricsRepoMock,
				RecordRepo:  &recordRepoMock,
			})

			nodeRepoMock.AssertNumberOfCalls(t, "AddNodeToActive", test.addToActiveNodesNumberOfCalls)
			nodeRepoMock.AssertNumberOfCalls(t, "Save", test.saveNodeNumberOfCalls)
		})
	}
}
