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

// AuthKeyEntry defines a single API key with user identity.
type AuthKeyEntry struct {
	Key       string `yaml:"key"`
	User      string `yaml:"user"`
	Group     string `yaml:"group"`
	RateLimit int    `yaml:"rate_limit"` // requests per minute
}

// AuthConfig controls virtual API key authentication and rate limiting.
type AuthConfig struct {
	Enabled   bool           `yaml:"enabled"`
	Keys      []AuthKeyEntry `yaml:"keys"`
	InjectKey string         `yaml:"inject_key"` // real upstream API key to inject
}

// CacheConfig controls response caching.
type CacheConfig struct {
	Enabled    bool `yaml:"enabled"`
	TTLSeconds int  `yaml:"ttl_seconds"`
	MaxEntries int  `yaml:"max_entries"`
}

// RetryConfig controls auto-retry and fallback routing.
type RetryConfig struct {
	Enabled         bool     `yaml:"enabled"`
	MaxRetries      int      `yaml:"max_retries"`
	BackoffMs       int      `yaml:"backoff_ms"`
	FallbackTargets []string `yaml:"fallback_targets"`
}

// SpendPricing defines cost per 1M tokens.
type SpendPricing struct {
	Input  float64 `yaml:"input"`
	Output float64 `yaml:"output"`
}

// SpendConfig controls token tracking and budget enforcement.
type SpendConfig struct {
	Enabled bool               `yaml:"enabled"`
	Budgets map[string]float64 `yaml:"budgets"` // group -> max USD
	Pricing SpendPricing       `yaml:"pricing"`
}

// DLPConfig controls data loss prevention policies.
type DLPConfig struct {
	Policies map[string]string `yaml:"policies"` // PIIType -> "redact" or "block"
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
	Auth    AuthConfig   `yaml:"auth"`
	Cache   CacheConfig  `yaml:"cache"`
	Retry   RetryConfig  `yaml:"retry"`
	Spend   SpendConfig  `yaml:"spend"`
	DLP     DLPConfig    `yaml:"dlp"`
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
		Cache: CacheConfig{
			TTLSeconds: 300,
			MaxEntries: 1000,
		},
		Retry: RetryConfig{
			MaxRetries: 3,
			BackoffMs:  500,
		},
		Spend: SpendConfig{
			Pricing: SpendPricing{
				Input:  3.00,
				Output: 15.00,
			},
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
	if cfg.Cache.TTLSeconds == 0 {
		cfg.Cache.TTLSeconds = 300
	}
	if cfg.Cache.MaxEntries == 0 {
		cfg.Cache.MaxEntries = 1000
	}
	if cfg.Retry.MaxRetries == 0 {
		cfg.Retry.MaxRetries = 3
	}
	if cfg.Retry.BackoffMs == 0 {
		cfg.Retry.BackoffMs = 500
	}
	if cfg.Spend.Pricing.Input == 0 {
		cfg.Spend.Pricing.Input = 3.00
	}
	if cfg.Spend.Pricing.Output == 0 {
		cfg.Spend.Pricing.Output = 15.00
	}

	return cfg, nil
}
