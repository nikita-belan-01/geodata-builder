package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (r *ListFilesRepository) WriteParts(dir string, chunks [][]string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	active := map[string]bool{}

	for i, domains := range chunks {
		if len(domains) == 0 {
			continue
		}

		target := filepath.Join(dir, fmt.Sprintf("part_%d.txt", i+1))
		active[target] = true

		if err := writeFileAtomically(dir, target, []byte(strings.Join(domains, "\n"))); err != nil {
			return err
		}
	}

	stale, err := filepath.Glob(filepath.Join(dir, "part_*.txt"))
	if err != nil {
		return fmt.Errorf("list stale parts: %w", err)
	}

	for _, path := range stale {
		if active[path] {
			continue
		}

		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove stale part %s: %w", path, err)
		}
	}

	r.log.Debug("parts written", "dir", dir, "parts", len(active))

	return nil
}

func (r *ListFilesRepository) WriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create dir %s: %w", dir, err)
	}

	return writeFileAtomically(dir, path, data)
}

func writeFileAtomically(dir, target string, data []byte) error {
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(target)+".*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)

		return fmt.Errorf("write temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)

		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, target); err != nil {
		os.Remove(tmpPath)

		return fmt.Errorf("replace %s: %w", target, err)
	}

	return nil
}
