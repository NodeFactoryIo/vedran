package record

//// FailedRequest should be called when rpc response is invalid to penalize node.
//// It does not return value as it should be called in separate goroutine
//func FailedRequest(node models.Node, repositories repositories.Repos) {
//	actions.PenalizeNode(node, repositories)
//
//	err := repositories.RecordRepo.Save(&models.Record{
//		NodeId:    node.ID,
//		Timestamp: time.Now(),
//		Status:    "failed",
//	})
//	if err != nil {
//		log.Errorf("Failed saving failed request because of: %v", err)
//	}
//}
//
//// SuccessfulRequest should be called when rpc response is valid to reward node.
//// It does not return value as it should be called in separate goroutine
//func SuccessfulRequest(node models.Node, repositories repositories.Repos) {
//	repositories.NodeRepo.RewardNode(node)
//
//	err := repositories.RecordRepo.Save(&models.Record{
//		NodeId:    node.ID,
//		Timestamp: time.Now(),
//		Status:    "successful",
//	})
//	if err != nil {
//		log.Errorf("Failed saving successful request because of: %v", err)
//	}
//}
