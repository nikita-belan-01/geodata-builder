package service

import (
	"context"

	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
	"github.com/nikita-belan-01/netcraze_route_builder/pkg/logger"
)

type ResolverRepository interface {
	ResolveSites(ctx context.Context, domains []string) []domain.AmneziaSite
}

type AmneziaService struct {
	resolver ResolverRepository
	log      *logger.Logger
}

func NewAmneziaService(resolver ResolverRepository, log *logger.Logger) *AmneziaService {
	if log == nil {
		log = logger.Discard()
	}

	return &AmneziaService{
		resolver: resolver,
		log:      log,
	}
}
