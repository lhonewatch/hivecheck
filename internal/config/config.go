// Package config provides configuration loading and validation for hivecheck.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// CheckConfig defines a single health check target.
type CheckConfig struct {
	Name     string        `yaml:"name"`
	URL      string        `yaml:"url"`
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
	WarnAt   float64       `yaml:"warn_at"`
	CritAt   float64       `yaml:"crit_at"`
}

// AlertConfig defines an alerting webhook endpoint.
type AlertConfig struct {
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
}

// Config is the top-level hivecheck configuration.
type Config struct {
	Checks []CheckConfig `yaml:"checks"`
	Alerts []AlertConfig `yaml:"alerts"`
}

// Load reads and parses a YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks that the configuration is semantically valid.
func (c *Config) Validate() error {
	if len(c.Checks) == 0 {
		return fmt.Errorf("config: at least one check must be defined")
	}
	for i, ch := range c.Checks {
		if ch.Name == "" {
			return fmt.Errorf("config: check[%d]: name is required", i)
		}
		if ch.URL == "" {
			return fmt.Errorf("config: check[%d] %q: url is required", i, ch.Name)
		}
		if ch.WarnAt > ch.CritAt {
			return fmt.Errorf("config: check[%d] %q: warn_at must be <= crit_at", i, ch.Name)
		}
	}
	return nil
}
