package service

import (
	"context"

	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
)

func (s *AmneziaService) Sites(
	ctx context.Context,
	domains []string,
) (sites []domain.AmneziaSite, resolved int) {
	sites = s.resolver.ResolveSites(ctx, domains)

	unique := make(map[string]bool, len(sites))
	for _, site := range sites {
		unique[site.Hostname] = true
	}

	s.log.Info("amnezia sites resolved",
		"requested", len(domains),
		"resolved", len(unique),
		"entries", len(sites))

	return sites, len(unique)
}
