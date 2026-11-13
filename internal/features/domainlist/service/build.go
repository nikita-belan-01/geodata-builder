package service

import (
	"context"
	"fmt"

	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
)

func (s *DomainListService) BuildDomains(ctx context.Context, list domain.List) ([]string, error) {
	rules, err := s.source.FetchRules(ctx, list.Roots)
	if err != nil {
		return nil, fmt.Errorf("fetch rules: %w", err)
	}

	domains := domain.DomainsFromRules(rules, list.Optimized)

	before := len(domains)
	domains = domain.CollapseCoveredSubdomains(domains)

	s.log.Info("list built",
		"list", list.Name,
		"domains", len(domains),
		"collapsed", before-len(domains))

	if len(domains) == 0 {
		return nil, domain.NewInvalidArgumentError(
			fmt.Sprintf("list %s came out empty", list.Name), domain.ErrEmptyDomainList)
	}

	return domains, nil
}
