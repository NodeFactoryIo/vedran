package actions

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
)

type Actions interface {
	PenalizeNode(node models.Node, repositories repositories.Repos)
}

type actions struct{}

func NewActions() Actions {
	return &actions{}
}
