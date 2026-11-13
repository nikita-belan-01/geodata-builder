package service

import (
	"context"
	"fmt"

	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
)

func (s *DomainListService) BuildChunks(
	ctx context.Context,
	list domain.List,
	partCount, linesPerFile int,
) ([][]string, error) {
	domains, err := s.BuildDomains(ctx, list)
	if err != nil {
		return nil, err
	}

	chunks, err := domain.SplitIntoFixedParts(domains, partCount, linesPerFile)
	if err != nil {
		return nil, err
	}

	return chunks, nil
}

func (s *DomainListService) SplitDomains(domains []string, linesPerFile int) [][]string {
	return domain.SplitIntoParts(domains, linesPerFile)
}

func (s *DomainListService) LoadChunks(dir string, partCount, linesPerFile int) ([][]string, error) {
	chunks, err := s.storage.ReadParts(dir, partCount, linesPerFile)
	if err != nil {
		return nil, fmt.Errorf("read parts from %s: %w", dir, err)
	}

	return chunks, nil
}

func (s *DomainListService) SaveChunks(dir string, chunks [][]string) error {
	if err := s.storage.WriteParts(dir, chunks); err != nil {
		return fmt.Errorf("write parts to %s: %w", dir, err)
	}

	return nil
}
