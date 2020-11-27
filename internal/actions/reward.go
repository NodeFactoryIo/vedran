package actions

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
)

// RewardNode updates LastUsed for provided node
func (a actions) RewardNode(node models.Node, repositories repositories.Repos) {
	repositories.NodeRepo.UpdateNodeUsed(node)
}
