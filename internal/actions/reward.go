package actions

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
)

func (a actions) RewardNode(node models.Node, repositories repositories.Repos) {
	repositories.NodeRepo.RewardNode(node)
}
