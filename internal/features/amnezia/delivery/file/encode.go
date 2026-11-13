package file

import (
	"encoding/json"
	"fmt"

	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
)

func Encode(sites []domain.AmneziaSite) ([]byte, error) {
	data, err := json.MarshalIndent(newSiteDTOs(sites), "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal amnezia sites: %w", err)
	}

	return data, nil
}
