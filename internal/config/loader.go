package config

import (
	"os"

	"github.com/Prateet-Github/sentinel/internal/core"
	"gopkg.in/yaml.v3"
)

func Load(path string) (*core.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg core.Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
