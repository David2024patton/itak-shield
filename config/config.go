package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// CustomRule defines an organization-specific PII detection pattern.
type CustomRule struct {
	Name    string `yaml:"name"`
	Pattern string `yaml:"pattern"`
}

// AuditConfig controls the structured audit logger.
type AuditConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Path      string `yaml:"path"`
	MaxSizeMB int    `yaml:"max_size_mb"`
	MaxFiles  int    `yaml:"max_files"`
}

// RulesConfig controls which PII patterns are active.
type RulesConfig struct {
	Custom   []CustomRule `yaml:"custom"`
	Disabled []string     `yaml:"disabled"`
}

// HealthConfig controls the health check endpoint.
type HealthConfig struct {
	Enabled bool `yaml:"enabled"`
}

// Config is the top-level configuration for iTaK Shield.
// All fields are optional with sensible defaults.
type Config struct {
	Listen  string       `yaml:"listen"`
	Target  string       `yaml:"target"`
	Verbose bool         `yaml:"verbose"`
	Audit   AuditConfig  `yaml:"audit"`
	Rules   RulesConfig  `yaml:"rules"`
	Health  HealthConfig `yaml:"health"`
}

// Defaults returns a Config with sensible defaults for personal use.
func Defaults() *Config {
	return &Config{
		Listen:  "127.0.0.1",
		Verbose: false,
		Audit: AuditConfig{
			Enabled:   false,
			Path:      "audit.jsonl",
			MaxSizeMB: 100,
			MaxFiles:  10,
		},
		Health: HealthConfig{
			Enabled: true,
		},
	}
}

// Load reads a YAML config file and returns a Config.
// If the file doesn't exist, it returns defaults (no error).
// This makes the config file entirely optional for personal use.
func Load(path string) (*Config, error) {
	cfg := Defaults()

	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Apply defaults for zero values.
	if cfg.Listen == "" {
		cfg.Listen = "127.0.0.1"
	}
	if cfg.Audit.MaxSizeMB == 0 {
		cfg.Audit.MaxSizeMB = 100
	}
	if cfg.Audit.MaxFiles == 0 {
		cfg.Audit.MaxFiles = 10
	}
	if cfg.Audit.Path == "" {
		cfg.Audit.Path = "audit.jsonl"
	}

	return cfg, nil
}
