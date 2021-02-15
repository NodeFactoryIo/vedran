package penalize

import (
	"testing"
	"time"

	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	"github.com/NodeFactoryIo/vedran/internal/whitelist"
	repoMocks "github.com/NodeFactoryIo/vedran/mocks/repositories"
	"github.com/stretchr/testify/mock"
)

func TestScheduleCheckForPenalizedNode(t *testing.T) {
	tests := []struct {
		name                              string
		nodeID                            string
		node                              models.Node
		increaseNodeCooldownNumberOfCalls int
		addToActiveNode                   []models.Node
		addToActiveNodesNumberOfCalls     int
		setNodeAsInactiveNumberOfCalls    int
		resetNodeCooldownNumberOfCalls    int
		nodePing                          []*models.Ping
		nodeMetrics                       []*models.Metrics
		latestMetrics                     []*models.LatestBlockMetrics
		increaseNodeCooldown              []*models.Node
	}{
		{
			name:   "penalized node becomes active on first check",
			nodeID: "1",
			node: models.Node{
				ID:       "1",
				Cooldown: 1,
			},
			addToActiveNode: []models.Node{
				{
					ID:       "1",
					Cooldown: 0,
				},
			},
			addToActiveNodesNumberOfCalls:  1,
			setNodeAsInactiveNumberOfCalls: 0,
			resetNodeCooldownNumberOfCalls: 1,
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
					TargetBlockHeight:    1000,
				},
			},
			latestMetrics: []*models.LatestBlockMetrics{
				{
					BestBlockHeight:      1001,
					FinalizedBlockHeight: 998,
				},
			},
			increaseNodeCooldown:              nil,
			increaseNodeCooldownNumberOfCalls: 0,
		},
		{
			name:   "penalized node schedule one check",
			nodeID: "1",
			node: models.Node{
				ID:       "1",
				Cooldown: 1,
			},
			addToActiveNode: []models.Node{
				{
					ID:       "1",
					Cooldown: 0,
				},
			},
			addToActiveNodesNumberOfCalls:  1,
			setNodeAsInactiveNumberOfCalls: 0,
			resetNodeCooldownNumberOfCalls: 1,
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
					TargetBlockHeight:    900,
					Timestamp: time.Now(),
				},
				{
					NodeId:               "1",
					BestBlockHeight:      1000,
					FinalizedBlockHeight: 995,
					TargetBlockHeight:    1000,
					Timestamp: time.Now(),
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
			increaseNodeCooldown: []*models.Node{
				{
					ID:       "1",
					Cooldown: 2,
				},
			},
			increaseNodeCooldownNumberOfCalls: 1,
		},
		{
			name:   "penalized node schedule multiple checks",
			nodeID: "1",
			node: models.Node{
				ID:       "1",
				Cooldown: 1,
			},
			addToActiveNode: []models.Node{
				{
					ID:       "1",
					Cooldown: 0,
				},
			},
			addToActiveNodesNumberOfCalls:  1,
			setNodeAsInactiveNumberOfCalls: 0,
			resetNodeCooldownNumberOfCalls: 1,
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
					TargetBlockHeight:    900,
					Timestamp: time.Now(),
				},
				{
					NodeId:               "1",
					BestBlockHeight:      900,
					FinalizedBlockHeight: 898,
					TargetBlockHeight:    900,
					Timestamp: time.Now(),
				},
				{
					NodeId:               "1",
					BestBlockHeight:      900,
					FinalizedBlockHeight: 898,
					TargetBlockHeight:    900,
					Timestamp: time.Now(),
				},
				{
					NodeId:               "1",
					BestBlockHeight:      1000,
					FinalizedBlockHeight: 995,
					TargetBlockHeight:    1000,
					Timestamp: time.Now(),
				},
			},
			latestMetrics: []*models.LatestBlockMetrics{
				{
					BestBlockHeight:      1001,
					FinalizedBlockHeight: 998,
				},
			},
			increaseNodeCooldown: []*models.Node{
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
			increaseNodeCooldownNumberOfCalls: 3,
		},
		{
			name:   "penalized node hits max cooldown",
			nodeID: "1",
			node: models.Node{
				ID:       "1",
				Cooldown: 510,
			},
			addToActiveNode:                nil,
			addToActiveNodesNumberOfCalls:  0,
			setNodeAsInactiveNumberOfCalls: 1,
			resetNodeCooldownNumberOfCalls: 0,
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
					TargetBlockHeight:    900,
					Timestamp: time.Now(),
				},
			},
			latestMetrics: []*models.LatestBlockMetrics{
				{
					BestBlockHeight:      1001,
					FinalizedBlockHeight: 998,
				},
			},
			increaseNodeCooldown: []*models.Node{
				{
					ID:       "1",
					Cooldown: 1020,
				},
				{
					ID:       "1",
					Cooldown: 2040,
				},
			},
			increaseNodeCooldownNumberOfCalls: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _ = whitelist.InitWhitelisting([]string{test.nodeID}, "")
			nodeRepoMock := repoMocks.NodeRepository{}

			// is mocked function called in test
			if test.addToActiveNode != nil {
				if len(test.addToActiveNode) == 1 { // same return value
					nodeRepoMock.On("AddNodeToActive", test.addToActiveNode[0].ID).Return(nil)
				} else { // multiple return values
					for _, n := range test.addToActiveNode {
						nodeRepoMock.On("AddNodeToActive", n.ID).Return().Once()
					}
				}
			}

			nodeRepoMock.On("ResetNodeCooldown", test.nodeID).Return(&models.Node{
				ID:       test.nodeID,
				Cooldown: 0,
			}, nil)
			nodeRepoMock.On("Save", mock.Anything).Return(nil)

			// is mocked function called in test
			if test.increaseNodeCooldown != nil {
				if len(test.increaseNodeCooldown) == 1 {
					nodeRepoMock.On(
						"IncreaseNodeCooldown",
						test.nodeID,
					).Return(test.increaseNodeCooldown[0], nil)
				} else {
					for _, node := range test.increaseNodeCooldown {
						nodeRepoMock.On("IncreaseNodeCooldown", test.nodeID).Return(node, nil).Once()
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
			nodeRepoMock.AssertNumberOfCalls(t, "IncreaseNodeCooldown", test.increaseNodeCooldownNumberOfCalls)
			nodeRepoMock.AssertNumberOfCalls(t, "ResetNodeCooldown", test.resetNodeCooldownNumberOfCalls)
			nodeRepoMock.AssertNumberOfCalls(t, "Save", test.setNodeAsInactiveNumberOfCalls)
		})
	}
}
