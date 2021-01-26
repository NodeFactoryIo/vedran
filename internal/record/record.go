package record

import (
	"github.com/NodeFactoryIo/vedran/internal/actions"
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
	log "github.com/sirupsen/logrus"
	"time"
)

// FailedRequest should be called when rpc response is invalid to penalize node.
// It does not return value as it should be called in separate goroutine
func FailedRequest(node models.Node, repositories repositories.Repos, actions actions.Actions) {
	actions.PenalizeNode(node, repositories, "failed request")

	err := repositories.RecordRepo.Save(&models.Record{
		NodeId:    node.ID,
		Timestamp: time.Now(),
		Status:    "failed",
	})
	if err != nil {
		log.Errorf("Failed saving failed request because of: %v", err)
	}

	log.Debugf("Node %s failed to serve successful request", node.ID)
}

// SuccessfulRequest should be called when rpc response is valid to reward node.
// It does not return value as it should be called in separate goroutine
func SuccessfulRequest(node models.Node, repositories repositories.Repos) {
	repositories.NodeRepo.UpdateNodeUsed(node)

	err := repositories.RecordRepo.Save(&models.Record{
		NodeId:    node.ID,
		Timestamp: time.Now(),
		Status:    "successful",
	})
	if err != nil {
		log.Errorf("Failed saving successful request because of: %v", err)
	}

	log.Debugf("Node %s served successful request", node.ID)
}
