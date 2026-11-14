package config

import (
	"fmt"
	"os"
)

const DefaultPath = ".env"

type Config struct {
	Logger Logger
	Router Router
	Lists  Lists
}

func OptionalPath(path string) string {
	if path != DefaultPath {
		return path
	}

	if _, err := os.Stat(path); err != nil {
		return ""
	}

	return path
}

func New(confPath string) (*Config, error) {
	if err := loadEnv(confPath); err != nil {
		return nil, fmt.Errorf("load env by path %s: %w", confPath, err)
	}

	loggerConfig, err := newLogger()
	if err != nil {
		return nil, fmt.Errorf("init logger config: %w", err)
	}

	router, err := newRouter()
	if err != nil {
		return nil, fmt.Errorf("init router config: %w", err)
	}

	lists, err := newLists()
	if err != nil {
		return nil, fmt.Errorf("init lists config: %w", err)
	}

	return &Config{
		Logger: loggerConfig,
		Router: router,
		Lists:  lists,
	}, nil
}
