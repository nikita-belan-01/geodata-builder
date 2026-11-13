package v2fly

import (
	"github.com/nikita-belan-01/netcraze_route_builder/pkg/logger"
)

const RepoZipURL = "https://github.com/v2fly/domain-list-community/archive/refs/heads/master.zip"

type DomainListRepository struct {
	url string
	log *logger.Logger
}

func NewDomainListRepository(log *logger.Logger) *DomainListRepository {
	if log == nil {
		log = logger.Discard()
	}

	return &DomainListRepository{
		url: RepoZipURL,
		log: log,
	}
}
