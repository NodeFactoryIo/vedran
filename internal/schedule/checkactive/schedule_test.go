package checkactive

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	actionMocks "github.com/NodeFactoryIo/vedran/mocks/actions"
	repoMocks "github.com/NodeFactoryIo/vedran/mocks/repositories"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func Test_scheduledTask(t *testing.T) {
	tests := []struct {
		name                              string
		allActiveNodes                    *[]models.Node
		nodePing                          []*models.Ping
		nodeMetrics                       []*models.Metrics
		latestMetrics                     []*models.LatestBlockMetrics
		penalizedNodes                    []models.Node
		penalizedNodesNumberOfCalls       int
		removeNodeFromActiveArgs          []models.Node
		removeNodeFromActiveNumberOfCalls int
	}{
		{
			name: "all active nodes",
			allActiveNodes: &[]models.Node{
				{
					ID: "1",
				},
				{
					ID: "2",
				},
				{
					ID: "3",
				},
			},
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
			penalizedNodes:              nil,
			penalizedNodesNumberOfCalls: 0,
		},
		{
			name: "penalize nodes with bad ping",
			allActiveNodes: &[]models.Node{
				{
					ID: "1",
				},
				{
					ID: "2", // not active
				},
				{
					ID: "3",
				},
				{
					ID: "4", // not active
				},
				{
					ID: "5",
				},
			},
			nodePing: []*models.Ping{
				{
					NodeId:    "1",
					Timestamp: time.Now(),
				},
				{
					NodeId:    "2",
					Timestamp: time.Now().Add(time.Duration(-1) * time.Hour), // before 1h
				},
				{
					NodeId:    "3",
					Timestamp: time.Now(),
				},
				{
					NodeId:    "4",
					Timestamp: time.Now().Add(time.Duration(-10) * time.Minute), // before 10 min
				},
				{
					NodeId:    "5",
					Timestamp: time.Now(),
				},
			},
			nodeMetrics: []*models.Metrics{
				{
					NodeId:               "1",
					BestBlockHeight:      1000,
					FinalizedBlockHeight: 995,
				},
				{
					NodeId:               "2",
					BestBlockHeight:      1000,
					FinalizedBlockHeight: 995,
				},
				{
					NodeId:               "3",
					BestBlockHeight:      1000,
					FinalizedBlockHeight: 995,
				},
				{
					NodeId:               "4",
					BestBlockHeight:      1000,
					FinalizedBlockHeight: 995,
				},
				{
					NodeId:               "5",
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
			penalizedNodes: []models.Node{
				{
					ID: "2",
				},
				{
					ID: "4",
				},
			},
			penalizedNodesNumberOfCalls: 2,
		},
		{
			name: "remove from active nodes with bad metrics",
			allActiveNodes: &[]models.Node{
				{
					ID: "1",
				},
				{
					ID: "2", // not active
				},
				{
					ID: "3",
				},
				{
					ID: "4", // not active
				},
				{
					ID: "5",
				},
			},
			nodePing: []*models.Ping{
				{
					NodeId:    "1",
					Timestamp: time.Now(),
				},
				{
					NodeId:    "2",
					Timestamp: time.Now(),
				},
				{
					NodeId:    "3",
					Timestamp: time.Now(),
				},
				{
					NodeId:    "4",
					Timestamp: time.Now(),
				},
				{
					NodeId:    "5",
					Timestamp: time.Now(),
				},
			},
			nodeMetrics: []*models.Metrics{
				{
					NodeId:               "1",
					BestBlockHeight:      1000,
					FinalizedBlockHeight: 995,
				},
				{
					NodeId:               "2",
					BestBlockHeight:      989,
					FinalizedBlockHeight: 986,
				},
				{
					NodeId:               "3",
					BestBlockHeight:      1000,
					FinalizedBlockHeight: 995,
				},
				{
					NodeId:               "4",
					BestBlockHeight:      989,
					FinalizedBlockHeight: 986,
				},
				{
					NodeId:               "5",
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
			penalizedNodes:              nil,
			penalizedNodesNumberOfCalls: 0,
			removeNodeFromActiveArgs: []models.Node{
				{
					ID: "2",
				},
				{
					ID: "4",
				},
			},
			removeNodeFromActiveNumberOfCalls: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			nodeRepoMock := repoMocks.NodeRepository{}
			nodeRepoMock.On("GetAllActiveNodes").Return(test.allActiveNodes).Once()

			if test.removeNodeFromActiveArgs != nil {
				if len(test.removeNodeFromActiveArgs) == 1 {
					nodeRepoMock.On("RemoveNodeFromActive", test.removeNodeFromActiveArgs[0].ID).Return(nil)
				} else {
					for _, node := range test.removeNodeFromActiveArgs {
						nodeRepoMock.On("RemoveNodeFromActive", node.ID).Return(nil).Once()
					}
				}
			}

			pingRepoMock := repoMocks.PingRepository{}
			if len(test.nodePing) == 1 { // same return value
				pingRepoMock.On("FindByNodeID", mock.Anything).Return(test.nodePing[0], nil)
			} else { // multiple return values
				for i, ping := range test.nodePing {
					pingRepoMock.On("FindByNodeID", (*test.allActiveNodes)[i].ID).Return(ping, nil).Once()
				}
			}

			metricsRepoMock := repoMocks.MetricsRepository{}
			if len(test.nodeMetrics) == 1 { // same return value
				metricsRepoMock.On("FindByID", mock.Anything).Return(test.nodeMetrics[0], nil)
			} else { // multiple return values
				for i, metric := range test.nodeMetrics {
					metricsRepoMock.On("FindByID", (*test.allActiveNodes)[i].ID).Return(metric, nil).Once()
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

			actionsMockObject := new(actionMocks.Actions)
			if test.penalizedNodes != nil {
				for _, pNode := range test.penalizedNodes {
					actionsMockObject.On("PenalizeNode", pNode, mock.Anything).Return().Once()
				}
			}

			scheduledTask(&repositories.Repos{
				NodeRepo:    &nodeRepoMock,
				PingRepo:    &pingRepoMock,
				MetricsRepo: &metricsRepoMock,
				RecordRepo:  &recordRepoMock,
			}, actionsMockObject)

			actionsMockObject.AssertNumberOfCalls(t, "PenalizeNode", test.penalizedNodesNumberOfCalls)
			nodeRepoMock.AssertNumberOfCalls(t, "RemoveNodeFromActive", test.removeNodeFromActiveNumberOfCalls)
			nodeRepoMock.AssertNumberOfCalls(t, "GetAllActiveNodes", 1)
		})
	}
}
