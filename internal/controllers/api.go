package controllers

import (
	"github.com/NodeFactoryIo/vedran/internal/repositories"
)

type ApiController struct {
	whitelistEnabled  bool
	repositories repositories.Repos
}

func NewApiController(
	whitelistEnabled bool,
	repositories repositories.Repos,
) *ApiController {
	return &ApiController{
		whitelistEnabled:  whitelistEnabled,
		repositories: repositories,
	}
}
