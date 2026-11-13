package fs

import (
	"regexp"

	"github.com/nikita-belan-01/netcraze_route_builder/pkg/logger"
)

var partFileRe = regexp.MustCompile(`^part_(\d+)\.txt$`)

type ListFilesRepository struct {
	log *logger.Logger
}

func NewListFilesRepository(log *logger.Logger) *ListFilesRepository {
	if log == nil {
		log = logger.Discard()
	}

	return &ListFilesRepository{log: log}
}
