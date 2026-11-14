package service

import (
	"context"

	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
	"github.com/nikita-belan-01/netcraze_route_builder/pkg/logger"
)

type RouterRepository interface {
	Sync(ctx context.Context, request domain.SyncRequest) (domain.SyncResult, error)
}

type RouterSyncService struct {
	router RouterRepository
	log    *logger.Logger
}

func NewRouterSyncService(router RouterRepository, log *logger.Logger) *RouterSyncService {
	if log == nil {
		log = logger.Discard()
	}

	return &RouterSyncService{
		router: router,
		log:    log,
	}
}
