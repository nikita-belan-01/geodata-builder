package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
)

type partFile struct {
	index int
	path  string
}

func (r *ListFilesRepository) ReadParts(dir string, partCount, linesPerFile int) ([][]string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "part_*.txt"))
	if err != nil {
		return nil, fmt.Errorf("list part files: %w", err)
	}

	var found []partFile

	for _, path := range matches {
		m := partFileRe.FindStringSubmatch(filepath.Base(path))
		if m == nil {
			continue
		}

		index, _ := strconv.Atoi(m[1])
		found = append(found, partFile{index: index, path: path})
	}

	sort.Slice(found, func(i, j int) bool { return found[i].index < found[j].index })

	if len(found) == 0 {
		return nil, domain.NewAppError(domain.KindNotFound,
			fmt.Sprintf("no part_N.txt files found in %s", dir), domain.ErrEmptyDomainList)
	}

	if len(found) > partCount {
		return nil, domain.NewInvalidArgumentError(
			fmt.Sprintf("found %d parts in %s, but %d are configured", len(found), dir, partCount),
			domain.ErrTooManyDomains)
	}

	chunks := make([][]string, 0, partCount)
	seen := map[string]bool{}

	for i, item := range found {
		expectedIndex := i + 1
		if item.index != expectedIndex {
			return nil, domain.NewInvalidArgumentError(fmt.Sprintf(
				"expected file part_%d.txt, but found %s; parts must be contiguous",
				expectedIndex, filepath.Base(item.path)), domain.ErrInvalidDomain)
		}

		domains, err := r.readDomainFile(item.path)
		if err != nil {
			return nil, err
		}

		if len(domains) > linesPerFile {
			return nil, domain.NewAppError(domain.KindTooLarge, fmt.Sprintf(
				"%s has %d lines, maximum is %d", item.path, len(domains), linesPerFile),
				domain.ErrTooManyDomains)
		}

		if err := checkDuplicates(item.path, domains, seen); err != nil {
			return nil, err
		}

		chunks = append(chunks, domains)
	}

	for len(chunks) < partCount {
		chunks = append(chunks, nil)
	}

	r.log.Debug("parts loaded", "dir", dir, "parts", len(found))

	return chunks, nil
}

func checkDuplicates(path string, domains []string, seen map[string]bool) error {
	local := make(map[string]bool, len(domains))

	for _, d := range domains {
		if seen[d] {
			return domain.NewInvalidArgumentError(
				fmt.Sprintf("domain repeats across parts: %s", d), domain.ErrInvalidDomain)
		}

		if local[d] {
			return domain.NewInvalidArgumentError(
				fmt.Sprintf("%s contains duplicate domains", path), domain.ErrInvalidDomain)
		}

		local[d] = true
	}

	for d := range local {
		seen[d] = true
	}

	return nil
}

func (r *ListFilesRepository) readDomainFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var domains []string

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if err := domain.ValidateDomain(line); err != nil {
			return nil, domain.NewInvalidArgumentError(fmt.Sprintf("invalid line in %s", path), err)
		}

		domains = append(domains, line)
	}

	return domains, nil
}
