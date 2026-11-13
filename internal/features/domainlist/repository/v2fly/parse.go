package v2fly

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
	"github.com/nikita-belan-01/netcraze_route_builder/pkg/logger"
)

func ParseLine(raw string) (rule string, attrs map[string]bool) {
	line := raw
	if idx := strings.Index(line, "#"); idx >= 0 {
		line = line[:idx]
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return "", nil
	}

	fields := strings.Fields(line)
	rule = strings.TrimSpace(fields[0])
	attrs = make(map[string]bool, len(fields)-1)

	for _, field := range fields[1:] {
		field = strings.TrimSpace(field)
		if strings.HasPrefix(field, "@") {
			attrs[field] = true
		}
	}

	return rule, attrs
}

type includeResolver struct {
	dataDir string
	visited map[string]bool
	log     *logger.Logger
}

func newIncludeResolver(dataDir string, log *logger.Logger) *includeResolver {
	return &includeResolver{
		dataDir: dataDir,
		visited: map[string]bool{},
		log:     log,
	}
}

func (r *includeResolver) resolve(name string) ([]domain.Rule, error) {
	if r.visited[name] {
		return nil, nil
	}

	r.visited[name] = true

	data, err := os.ReadFile(filepath.Join(r.dataDir, name))
	if err != nil {
		if os.IsNotExist(err) {
			r.log.Warn("include not found", "include", name)

			return nil, nil
		}

		return nil, err
	}

	var result []domain.Rule

	for _, raw := range strings.Split(string(data), "\n") {
		rule, attrs := ParseLine(strings.TrimRight(raw, "\r"))
		if rule == "" {
			continue
		}

		if nested, ok := strings.CutPrefix(rule, "include:"); ok {
			resolved, err := r.resolve(nested)
			if err != nil {
				return nil, err
			}

			result = append(result, resolved...)

			continue
		}

		result = append(result, domain.Rule{Name: rule, Attrs: attrs})
	}

	return result, nil
}
