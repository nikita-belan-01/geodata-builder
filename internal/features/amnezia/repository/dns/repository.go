package dns

import (
	"time"

	"github.com/nikita-belan-01/netcraze_route_builder/pkg/logger"
)

const (
	defaultLookupTimeout = 3 * time.Second
	defaultConcurrency   = 64
)

type ResolverRepository struct {
	timeout     time.Duration
	concurrency int
	log         *logger.Logger
}

func NewResolverRepository(log *logger.Logger) *ResolverRepository {
	if log == nil {
		log = logger.Discard()
	}

	return &ResolverRepository{
		timeout:     defaultLookupTimeout,
		concurrency: defaultConcurrency,
		log:         log,
	}
}
