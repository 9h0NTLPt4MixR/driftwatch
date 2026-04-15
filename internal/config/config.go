package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const defaultConfigFile = ".driftwatch.yaml"

// Service represents a single service declaration from version control.
type Service struct {
	Name     string            `yaml:"name"`
	Endpoint string            `yaml:"endpoint"`
	Env      map[string]string `yaml:"env"`
	Replicas int               `yaml:"replicas"`
	Image    string            `yaml:"image"`
}

// Config is the top-level configuration structure.
type Config struct {
	Version  string             `yaml:"version"`
	Services map[string]Service `yaml:"services"`
}

// Load reads and parses the driftwatch config file.
func Load(path string) (*Config, error) {
	if path == "" {
		path = defaultConfigFile
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if len(c.Services) == 0 {
		return fmt.Errorf("no services defined")
	}
	for name, svc := range c.Services {
		if svc.Endpoint == "" {
			return fmt.Errorf("service %q missing endpoint", name)
		}
	}
	return nil
}
