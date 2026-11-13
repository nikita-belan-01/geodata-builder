package service

import (
	"context"

	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
	"github.com/nikita-belan-01/netcraze_route_builder/pkg/logger"
)

type SourceRepository interface {
	FetchRules(ctx context.Context, roots []string) ([]domain.Rule, error)
}

type StorageRepository interface {
	ReadParts(dir string, partCount, linesPerFile int) ([][]string, error)
	WriteParts(dir string, chunks [][]string) error
}

type DomainListService struct {
	source  SourceRepository
	storage StorageRepository
	log     *logger.Logger
}

func NewDomainListService(
	source SourceRepository,
	storage StorageRepository,
	log *logger.Logger,
) *DomainListService {
	if log == nil {
		log = logger.Discard()
	}

	return &DomainListService{
		source:  source,
		storage: storage,
		log:     log,
	}
}
