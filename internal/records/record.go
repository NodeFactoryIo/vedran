package record

import (
	"time"

	"github.com/NodeFactoryIo/vedran/internal/models"
	log "github.com/sirupsen/logrus"
)

func FailedRequest(node models.Node, nodeRepo models.NodeRepository, recordRepo models.RecordRepository) {
	nodeRepo.PenalizeNode(node)

	err := recordRepo.Save(&models.Record{
		NodeId:    node.ID,
		Timestamp: time.Now(),
		Status:    "failed",
	})
	if err != nil {
		log.Errorf("Failed saving failed request because of: %v", err)
	}
}

func SuccessfulRequest(node models.Node, nodeRepo models.NodeRepository, recordRepo models.RecordRepository) {
	nodeRepo.RewardNode(node)

	err := recordRepo.Save(&models.Record{
		NodeId:    node.ID,
		Timestamp: time.Now(),
		Status:    "successful",
	})
	if err != nil {
		log.Errorf("Failed saving successful request because of: %v", err)
	}
}
