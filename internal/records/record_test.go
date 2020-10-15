package record

// TODO
//func TestFailedRequest(t *testing.T) {
//
//	tests := []struct {
//		name                    string
//		penalizedNodeCallCount  int
//		saveNodeRecordCallCount int
//		saveNodeRecordResult    error
//	}{
//		{
//			name:                    "Calls penalize node and saves request record",
//			penalizedNodeCallCount:  1,
//			saveNodeRecordCallCount: 1,
//			saveNodeRecordResult:    nil},
//		{
//			name:                    "Logs error if save request record fails",
//			penalizedNodeCallCount:  1,
//			saveNodeRecordCallCount: 1,
//			saveNodeRecordResult:    fmt.Errorf("Error")},
//	}
//
//	for _, tt := range tests {
//
//		t.Run(tt.name, func(t *testing.T) {
//			nodeRepoMock := mocks.NodeRepository{}
//			recordRepoMock := mocks.RecordRepository{}
//			node := models.Node{
//				ID: "test-id",
//			}
//
//			nodeRepoMock.On("PenalizeNode", node, mock.Anything).Once().Return()
//			recordRepoMock.On("Save", mock.Anything).Once().Return(tt.saveNodeRecordResult)
//
//			FailedRequest(node, &nodeRepoMock, &recordRepoMock)
//
//			assert.True(t, nodeRepoMock.AssertNumberOfCalls(t, "PenalizeNode", tt.penalizedNodeCallCount))
//			assert.True(t, recordRepoMock.AssertNumberOfCalls(t, "Save", tt.saveNodeRecordCallCount))
//		})
//
//	}
//}
//func TestSuccessfulRequest(t *testing.T) {
//	tests := []struct {
//		name                    string
//		rewardNodeCallCount     int
//		saveNodeRecordCallCount int
//		saveNodeRecordResult    error
//	}{
//		{
//			name:                    "Calls reward node and saves request record",
//			rewardNodeCallCount:     1,
//			saveNodeRecordCallCount: 1,
//			saveNodeRecordResult:    nil},
//		{
//			name:                    "Logs error if save request record fails",
//			rewardNodeCallCount:     1,
//			saveNodeRecordCallCount: 1,
//			saveNodeRecordResult:    fmt.Errorf("Error")},
//	}
//
//	for _, tt := range tests {
//
//		t.Run(tt.name, func(t *testing.T) {
//			nodeRepoMock := mocks.NodeRepository{}
//			recordRepoMock := mocks.RecordRepository{}
//			node := models.Node{
//				ID: "test-id",
//			}
//
//			nodeRepoMock.On("RewardNode", node).Once().Return()
//			recordRepoMock.On("Save", mock.Anything).Once().Return(tt.saveNodeRecordResult)
//
//			SuccessfulRequest(node, &nodeRepoMock, &recordRepoMock)
//
//			assert.True(t, nodeRepoMock.AssertNumberOfCalls(t, "RewardNode", tt.rewardNodeCallCount))
//			assert.True(t, recordRepoMock.AssertNumberOfCalls(t, "Save", tt.saveNodeRecordCallCount))
//		})
//
//	}
//}
