package config

import (
	"github.com/nikita-belan-01/netcraze_route_builder/pkg/default_env"
	"github.com/nikita-belan-01/netcraze_route_builder/pkg/logger"
)

type Logger struct {
	Dir    string
	Level  string
	Format string
}

func newLogger() (Logger, error) {
	level := default_env.GetString("LOG_LEVEL", "info")
	if err := logger.ValidateLogLevel(level); err != nil {
		return Logger{}, err
	}

	format := default_env.GetString("LOG_FORMAT", logger.FormatText)
	if err := logger.ValidateLogFormat(format); err != nil {
		return Logger{}, err
	}

	return Logger{
		Dir:    default_env.GetString("LOG_DIR", ""),
		Level:  level,
		Format: format,
	}, nil
}
