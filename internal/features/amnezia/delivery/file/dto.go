package file

import "github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"

type siteDTO struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
}

func newSiteDTOs(sites []domain.AmneziaSite) []siteDTO {
	dtos := make([]siteDTO, 0, len(sites))

	for _, site := range sites {
		dtos = append(dtos, siteDTO{
			Hostname: site.Hostname,
			IP:       site.IP,
		})
	}

	return dtos
}
