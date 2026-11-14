package netcrazecli

import (
	"context"
	"fmt"
	"slices"

	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/repository/netcraze"
)

const saveCommand = "system configuration save"

func (r *RouterRepository) Sync(
	ctx context.Context,
	request domain.SyncRequest,
) (domain.SyncResult, error) {
	session, err := netcraze.Dial(r.session, r.log)
	if err != nil {
		return domain.SyncResult{}, err
	}
	defer func() {
		if err := session.Close(); err != nil {
			r.log.Debug("close router session", "error", err.Error())
		}
	}()

	preflight, err := session.SendRaw("show running-config")
	if err != nil {
		return domain.SyncResult{}, err
	}

	if err := netcraze.TranslateResponse("show running-config", preflight); err != nil {
		return domain.SyncResult{}, err
	}

	policy := request.Policy

	if request.SelectPolicy != nil {
		discovered := parsePreflight(preflight, request.Prefix, domain.DefaultPolicy).policies

		policy, err = request.SelectPolicy(discovered)
		if err != nil {
			return domain.SyncResult{}, err
		}
	}

	info := parsePreflight(preflight, request.Prefix, policy)

	partCount := max(request.PartCount, highestManagedIndex(info, request.Prefix))
	commands := slices.Concat(request.Commands, clearCommands(request.Prefix, request.PartCount, partCount))

	mg, err := resolveManagedGroups(len(preflight), info, request.Prefix, partCount,
		request.RouteGateway, request.RouteInterface)
	if err != nil {
		return domain.SyncResult{}, err
	}

	r.log.Info("router preflight done",
		"policy", domain.PolicyLabel(policy),
		"parts", request.PartCount,
		"stale_parts_to_clear", partCount-request.PartCount,
		"missing_groups", mg.missingGroups,
		"missing_routes", len(mg.missingRoutes))

	saveRequested, err := r.replayCommands(ctx, session, commands, mg)
	if err != nil {
		return domain.SyncResult{}, err
	}

	if err := r.createRoutes(session, mg, policy); err != nil {
		return domain.SyncResult{}, err
	}

	if saveRequested {
		if _, err := session.Send(saveCommand); err != nil {
			return domain.SyncResult{}, err
		}

		r.log.Info("router configuration saved")
	}

	if err := session.Quit(); err != nil {
		return domain.SyncResult{}, err
	}

	return domain.SyncResult{
		Parts:         request.PartCount,
		ClearedParts:  partCount - request.PartCount,
		MissingGroups: mg.missingGroups,
		MissingRoutes: len(mg.missingRoutes),
		Policy:        policy,
	}, nil
}

func (r *RouterRepository) replayCommands(
	ctx context.Context,
	session *netcraze.Session,
	commands []string,
	mg managedGroups,
) (saveRequested bool, err error) {
	for lineNumber, command := range commands {
		if err := ctx.Err(); err != nil {
			return false, err
		}

		if command == "" {
			continue
		}

		if command == saveCommand {
			saveRequested = true

			continue
		}

		if _, err := session.Send(rewriteObjectGroupCommand(command, mg.displayToActual)); err != nil {
			return false, fmt.Errorf("line %d (%s); configuration not saved: %w", lineNumber+1, command, err)
		}
	}

	return saveRequested, nil
}

func (r *RouterRepository) createRoutes(
	session *netcraze.Session,
	mg managedGroups,
	policy string,
) error {
	if len(mg.missingRoutes) == 0 {
		return nil
	}

	inPolicy := policy != domain.DefaultPolicy

	if inPolicy {
		if _, err := session.Send("ip policy " + policy); err != nil {
			return fmt.Errorf("enter policy %s: %w", policy, err)
		}
	}

	for _, name := range mg.missingRoutes {
		command := fmt.Sprintf("dns-proxy route object-group %s %s", name, mg.routeSuffix)

		if _, err := session.Send(command); err != nil {
			return fmt.Errorf("create route %s; configuration not saved: %w", name, err)
		}

		r.log.Info("route created", "group", name, "policy", domain.PolicyLabel(policy))
	}

	if inPolicy {
		if _, err := session.Send("exit"); err != nil {
			return fmt.Errorf("leave policy %s: %w", policy, err)
		}
	}

	return nil
}
