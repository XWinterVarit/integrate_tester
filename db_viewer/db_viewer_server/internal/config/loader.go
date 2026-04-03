package config

import (
	"fmt"
	"os"

	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/model"
	"gopkg.in/yaml.v3"
)

func Load(path string) (*model.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg model.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.CORSOrigin == "" {
		cfg.Server.CORSOrigin = "http://localhost:3000"
	}

	return &cfg, nil
}
