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

			recordRepoMock := mocks.RecordRepository{}
			recordRepoMock.On("Save", mock.Anything).Once().Return(tt.saveNodeRecordResult)

			actionsMock := aMock.Actions{}
			actionsMock.On("PenalizeNode", node, mock.Anything, mock.Anything).Return()

			FailedRequest(node, repositories.Repos{
				RecordRepo: &recordRepoMock,
			}, &actionsMock)

			actionsMock.AssertNumberOfCalls(t, "PenalizeNode", tt.penalizedNodeCallCount)
			recordRepoMock.AssertNumberOfCalls(t, "Save", tt.saveNodeRecordCallCount)
		})

	}
}
func TestSuccessfulRequest(t *testing.T) {
	tests := []struct {
		name                    string
		updateNodeUsedCallCount int
		saveNodeRecordCallCount int
		saveNodeRecordResult    error
	}{
		{
			name:                    "Calls reward node and saves request record",
			updateNodeUsedCallCount: 1,
			saveNodeRecordCallCount: 1,
			saveNodeRecordResult:    nil},
		{
			name:                    "Logs error if save request record fails",
			updateNodeUsedCallCount: 1,
			saveNodeRecordCallCount: 1,
			saveNodeRecordResult:    fmt.Errorf("Error")},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			node := models.Node{
				ID: "test-id",
			}

			nodeRepoMock := mocks.NodeRepository{}
			nodeRepoMock.On("UpdateNodeUsed", node).Return()

			recordRepoMock := mocks.RecordRepository{}
			recordRepoMock.On("Save", mock.Anything).Once().Return(tt.saveNodeRecordResult)

			SuccessfulRequest(node, repositories.Repos{
				NodeRepo:   &nodeRepoMock,
				RecordRepo: &recordRepoMock,
			})

			recordRepoMock.AssertNumberOfCalls(t, "Save", tt.saveNodeRecordCallCount)
			nodeRepoMock.AssertNumberOfCalls(t, "UpdateNodeUsed", tt.updateNodeUsedCallCount)
		})

	}
}
