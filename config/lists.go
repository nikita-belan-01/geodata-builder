package config

import (
	"fmt"

	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
	"github.com/nikita-belan-01/netcraze_route_builder/pkg/default_env"
)

const (
	defaultOutputDir   = "netcraze_output"
	defaultArchiveName = "netcraze_output.zip"
	defaultListName    = "full"
	defaultSourceURL   = "https://github.com/v2fly/domain-list-community/archive/refs/heads/master.zip"
)

type Lists struct {
	Name         string
	SourceURL    string
	OutputDir    string
	ArchivePath  string
	LinesPerFile int
	UseExisting  bool
}

func newLists() (Lists, error) {
	linesPerFile, err := default_env.GetNumber("NETCRAZE_LINES_PER_FILE", domain.DefaultLinesPerFile)
	if err != nil {
		return Lists{}, err
	}

	if err := validateSize(linesPerFile); err != nil {
		return Lists{}, err
	}

	useExisting, err := default_env.GetBool("NETCRAZE_USE_EXISTING", false)
	if err != nil {
		return Lists{}, err
	}

	name := default_env.GetString("NETCRAZE_LIST", defaultListName)
	if _, err := domain.LookupList(name); err != nil {
		return Lists{}, err
	}

	sourceURL := default_env.GetString("NETCRAZE_SOURCE_URL", defaultSourceURL)
	if err := validateHTTPURL(sourceURL); err != nil {
		return Lists{}, fmt.Errorf("NETCRAZE_SOURCE_URL: %w", err)
	}

	return Lists{
		Name:         name,
		SourceURL:    sourceURL,
		OutputDir:    default_env.GetString("NETCRAZE_OUTPUT_DIR", defaultOutputDir),
		ArchivePath:  default_env.GetString("NETCRAZE_ARCHIVE", defaultArchiveName),
		LinesPerFile: linesPerFile,
		UseExisting:  useExisting,
	}, nil
}
