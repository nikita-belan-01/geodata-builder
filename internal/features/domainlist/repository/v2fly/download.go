package v2fly

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
)

func (r *DomainListRepository) downloadDataDir(ctx context.Context, workDir string) (string, error) {
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return "", fmt.Errorf("create work dir: %w", err)
	}

	archivePath := filepath.Join(workDir, "domain-list-community.zip")

	r.log.Info("downloading domain-list-community", "url", r.url)

	if err := r.download(ctx, r.url, archivePath); err != nil {
		return "", domain.NewUnavailableError("download domain-list-community", err)
	}

	r.log.Debug("extracting archive", "path", archivePath)

	if err := extractZip(archivePath, workDir); err != nil {
		return "", fmt.Errorf("extract archive: %w", err)
	}

	matches, err := filepath.Glob(filepath.Join(workDir, "domain-list-community-*", "data"))
	if err != nil {
		return "", fmt.Errorf("look for the data directory: %w", err)
	}

	if len(matches) == 0 {
		return "", domain.NewAppError(domain.KindNotFound,
			"data directory not found in the domain-list-community archive", domain.ErrSourceUnavailable)
	}

	return matches[0], nil
}

func (r *DomainListRepository) download(ctx context.Context, url, destination string) error {
	httpErr := downloadWithHTTP(ctx, url, destination)
	if httpErr == nil {
		return nil
	}

	curlPath, lookErr := exec.LookPath("curl")
	if lookErr != nil {
		return httpErr
	}

	r.log.Warn("HTTP download failed, retrying with curl", "error", httpErr.Error())

	cmd := exec.CommandContext(ctx, curlPath,
		"--fail", "--location", "--silent", "--show-error", "--output", destination, url)
	if runErr := cmd.Run(); runErr != nil {
		return httpErr
	}

	return nil
}

func downloadWithHTTP(ctx context.Context, url, destination string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", response.Status)
	}

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, response.Body)

	return err
}

func extractZip(archivePath, destDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	destDir, err = filepath.Abs(destDir)
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		targetPath, err := safeJoin(destDir, file.Name)
		if err != nil {
			return err
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return err
			}

			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}

		if err := extractZipFile(file, targetPath); err != nil {
			return err
		}
	}

	return nil
}

func safeJoin(destDir, name string) (string, error) {
	target := filepath.Join(destDir, name)
	if target != destDir && !strings.HasPrefix(target, destDir+string(os.PathSeparator)) {
		return "", fmt.Errorf("unsafe path in archive: %q", name)
	}

	return target, nil
}

func extractZipFile(file *zip.File, targetPath string) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, rc); err != nil {
		return err
	}

	return err
}
