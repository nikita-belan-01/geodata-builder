package v2fly

import (
	"github.com/nikita-belan-01/netcraze_route_builder/pkg/logger"
)

type DomainListRepository struct {
	url string
	log *logger.Logger
}

func NewDomainListRepository(url string, log *logger.Logger) *DomainListRepository {
	if log == nil {
		log = logger.Discard()
	}

	return &DomainListRepository{
		url: url,
		log: log,
	}
}
