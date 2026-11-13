package v2fly

import (
	"context"
	"fmt"
	"os"

	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
)

func (r *DomainListRepository) FetchRules(ctx context.Context, roots []string) ([]domain.Rule, error) {
	workDir, err := os.MkdirTemp("", "netcraze-domainlist-")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	defer func() {
		if err := os.RemoveAll(workDir); err != nil {
			r.log.Warn("remove temp dir", "dir", workDir, "error", err.Error())
		}
	}()

	dataDir, err := r.downloadDataDir(ctx, workDir)
	if err != nil {
		return nil, err
	}

	resolver := newIncludeResolver(dataDir, r.log)

	var rules []domain.Rule

	for _, root := range roots {
		resolved, err := resolver.resolve(root)
		if err != nil {
			return nil, fmt.Errorf("resolve root %s: %w", root, err)
		}

		rules = append(rules, resolved...)
	}

	r.log.Debug("rules fetched", "roots", len(roots), "rules", len(rules))

	return rules, nil
}
