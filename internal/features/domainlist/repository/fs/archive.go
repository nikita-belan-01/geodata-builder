package fs

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func (r *ListFilesRepository) Archive(sourceDir, zipPath string) error {
	if err := os.Remove(zipPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove old archive: %w", err)
	}

	out, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("create archive: %w", err)
	}
	writer := zip.NewWriter(out)

	discard := func() {
		r.closeOnFailure(out, zipPath)
		r.removeLeftover(zipPath)
	}

	if err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relative, err := filepath.Rel(filepath.Dir(sourceDir), path)
		if err != nil {
			return err
		}

		return r.addFileToZip(writer, path, relative)
	}); err != nil {
		r.closeOnFailure(writer, zipPath)
		discard()

		return fmt.Errorf("pack %s: %w", sourceDir, err)
	}

	if err := writer.Close(); err != nil {
		discard()

		return fmt.Errorf("finalize archive: %w", err)
	}

	if err := out.Close(); err != nil {
		r.removeLeftover(zipPath)

		return fmt.Errorf("close archive file: %w", err)
	}

	r.log.Info("archive ready", "path", zipPath)

	return nil
}

func (r *ListFilesRepository) addFileToZip(writer *zip.Writer, path, name string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			r.log.Debug("close source file", "path", path, "error", err.Error())
		}
	}()

	entry, err := writer.Create(name)
	if err != nil {
		return err
	}

	if _, err = io.Copy(entry, file); err != nil {
		return err
	}

	return nil
}
