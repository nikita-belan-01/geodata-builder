package netcrazecli

import (
	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/repository/netcraze"
	"github.com/nikita-belan-01/netcraze_route_builder/pkg/logger"
)

type RouterRepository struct {
	session netcraze.SessionConfig
	log     *logger.Logger
}

func NewRouterRepository(session netcraze.SessionConfig, log *logger.Logger) *RouterRepository {
	if log == nil {
		log = logger.Discard()
	}

	return &RouterRepository{
		session: session,
		log:     log,
	}
}
