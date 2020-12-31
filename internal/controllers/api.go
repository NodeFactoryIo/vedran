package controllers

import (
	"github.com/NodeFactoryIo/vedran/internal/actions"
	"github.com/NodeFactoryIo/vedran/internal/repositories"
)

type ApiController struct {
	whitelistEnabled bool
	repositories     repositories.Repos
	actions          actions.Actions
}

func NewApiController(
	whitelistEnabled bool,
	repositories repositories.Repos,
	actions actions.Actions,
) *ApiController {
	return &ApiController{
		whitelistEnabled: whitelistEnabled,
		repositories:     repositories,
		actions:          actions,
	}
}
