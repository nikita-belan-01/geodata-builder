package service

import (
	"context"

	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
)

type SyncParams struct {
	Prefix         string
	Policy         string
	RouteGateway   string
	RouteInterface string
	SelectPolicy   func(policies []domain.Policy) (string, error)
}

func (s *RouterSyncService) Sync(
	ctx context.Context,
	chunks [][]string,
	params SyncParams,
) (domain.SyncResult, error) {
	total, nonEmpty := domain.CountDomains(chunks)

	s.log.Info("syncing router",
		"prefix", params.Prefix,
		"parts", len(chunks),
		"domains", total,
		"non_empty_parts", nonEmpty)

	result, err := s.router.Sync(ctx, domain.SyncRequest{
		Prefix:         params.Prefix,
		PartCount:      len(chunks),
		Commands:       BuildCommands(chunks, params.Prefix),
		Policy:         params.Policy,
		RouteGateway:   params.RouteGateway,
		RouteInterface: params.RouteInterface,
		SelectPolicy:   params.SelectPolicy,
	})
	if err != nil {
		return domain.SyncResult{}, err
	}

	s.log.Info("router synced",
		"policy", domain.PolicyLabel(result.Policy),
		"parts", result.Parts,
		"cleared_parts", result.ClearedParts,
		"missing_groups", result.MissingGroups,
		"missing_routes", result.MissingRoutes)

	return result, nil
}
