package record

import (
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	aMock "github.com/NodeFactoryIo/vedran/mocks/actions"
	mocks "github.com/NodeFactoryIo/vedran/mocks/repositories"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestFailedRequest(t *testing.T) {

	tests := []struct {
		name                    string
		penalizedNodeCallCount  int
		saveNodeRecordCallCount int
		saveNodeRecordResult    error
	}{
		{
			name:                    "Calls penalize node and saves request record",
			penalizedNodeCallCount:  1,
			saveNodeRecordCallCount: 1,
			saveNodeRecordResult:    nil},
		{
			name:                    "Logs error if save request record fails",
			penalizedNodeCallCount:  1,
			saveNodeRecordCallCount: 1,
			saveNodeRecordResult:    fmt.Errorf("error")},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			node := models.Node{
				ID: "test-id",
			}

			nodeRepoMock := mocks.NodeRepository{}
			pingRepoMock := mocks.PingRepository{}
			metricsRepoMock := mocks.MetricsRepository{}
			recordRepoMock := mocks.RecordRepository{}
			recordRepoMock.On("Save", mock.Anything).Once().Return(tt.saveNodeRecordResult)

			actionsMock := aMock.Actions{}
			actionsMock.On("PenalizeNode", node, mock.Anything).Return()

			FailedRequest(node, repositories.Repos{
				NodeRepo:    &nodeRepoMock,
				PingRepo:    &pingRepoMock,
				MetricsRepo: &metricsRepoMock,
				RecordRepo:  &recordRepoMock,
			}, &actionsMock)

			actionsMock.AssertNumberOfCalls(t, "PenalizeNode", tt.penalizedNodeCallCount)
			recordRepoMock.AssertNumberOfCalls(t, "Save", tt.saveNodeRecordCallCount)
		})

	}
}
func TestSuccessfulRequest(t *testing.T) {
	tests := []struct {
		name                    string
		rewardNodeCallCount     int
		saveNodeRecordCallCount int
		saveNodeRecordResult    error
	}{
		{
			name:                    "Calls reward node and saves request record",
			rewardNodeCallCount:     1,
			saveNodeRecordCallCount: 1,
			saveNodeRecordResult:    nil},
		{
			name:                    "Logs error if save request record fails",
			rewardNodeCallCount:     1,
			saveNodeRecordCallCount: 1,
			saveNodeRecordResult:    fmt.Errorf("Error")},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			nodeRepoMock := mocks.NodeRepository{}
			pingRepoMock := mocks.PingRepository{}
			metricsRepoMock := mocks.MetricsRepository{}
			recordRepoMock := mocks.RecordRepository{}

			node := models.Node{
				ID: "test-id",
			}
			recordRepoMock.On("Save", mock.Anything).Once().Return(tt.saveNodeRecordResult)

			actionsMock := aMock.Actions{}
			actionsMock.On("RewardNode", node, mock.Anything).Return()

			SuccessfulRequest(node, repositories.Repos{
				NodeRepo:    &nodeRepoMock,
				PingRepo:    &pingRepoMock,
				MetricsRepo: &metricsRepoMock,
				RecordRepo:  &recordRepoMock,
			}, &actionsMock)

			recordRepoMock.AssertNumberOfCalls(t, "Save", tt.saveNodeRecordCallCount)
			actionsMock.AssertNumberOfCalls(t, "RewardNode", tt.rewardNodeCallCount)
		})

	}
}