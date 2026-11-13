package config

import (
	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
	"github.com/nikita-belan-01/netcraze_route_builder/pkg/default_env"
)

const (
	defaultOutputDir   = "netcraze_output"
	defaultArchiveName = "netcraze_output.zip"
	defaultListName    = "full"
)

type Lists struct {
	Name         string
	OutputDir    string
	ArchivePath  string
	LinesPerFile int
	PartCount    int
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

	partCount, err := default_env.GetNumber("NETCRAZE_PARTS", domain.DefaultPartCount)
	if err != nil {
		return Lists{}, err
	}

	if err := validateSize(partCount); err != nil {
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

	return Lists{
		Name:         name,
		OutputDir:    default_env.GetString("NETCRAZE_OUTPUT_DIR", defaultOutputDir),
		ArchivePath:  default_env.GetString("NETCRAZE_ARCHIVE", defaultArchiveName),
		LinesPerFile: linesPerFile,
		PartCount:    partCount,
		UseExisting:  useExisting,
	}, nil
}
