package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()
	if cfg.Listen != "127.0.0.1" {
		t.Errorf("Listen: got %q, want 127.0.0.1", cfg.Listen)
	}
	if cfg.Verbose {
		t.Error("Verbose should default to false")
	}
	if cfg.Audit.Enabled {
		t.Error("Audit should default to disabled")
	}
	if cfg.Health.Enabled != true {
		t.Error("Health should default to enabled")
	}
	t.Logf("Defaults: PASSED")
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load("nonexistent.yaml")
	if err != nil {
		t.Fatalf("Load should not error on missing file: %v", err)
	}
	if cfg.Listen != "127.0.0.1" {
		t.Errorf("Expected defaults on missing file, got Listen=%q", cfg.Listen)
	}
	t.Logf("Load missing file: PASSED (returns defaults)")
}

func TestLoadEmptyPath(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load should not error on empty path: %v", err)
	}
	if cfg.Listen != "127.0.0.1" {
		t.Errorf("Expected defaults on empty path, got Listen=%q", cfg.Listen)
	}
	t.Logf("Load empty path: PASSED (returns defaults)")
}

func TestLoadValidConfig(t *testing.T) {
	content := `
listen: 0.0.0.0
target: https://api.openai.com
verbose: true

audit:
  enabled: true
  path: /var/log/audit.jsonl
  max_size_mb: 50
  max_files: 5

rules:
  custom:
    - name: EMPLOYEE_ID
      pattern: 'EMP-\d{6}'
    - name: PROJECT
      pattern: 'PROJ-[A-Z]{3}-\d{4}'
  disabled:
    - PATH
    - IP_ADDR

health:
  enabled: true
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "shield.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Listen != "0.0.0.0" {
		t.Errorf("Listen: got %q, want 0.0.0.0", cfg.Listen)
	}
	if cfg.Target != "https://api.openai.com" {
		t.Errorf("Target: got %q", cfg.Target)
	}
	if !cfg.Verbose {
		t.Error("Verbose should be true")
	}
	if !cfg.Audit.Enabled {
		t.Error("Audit should be enabled")
	}
	if cfg.Audit.MaxSizeMB != 50 {
		t.Errorf("Audit MaxSizeMB: got %d, want 50", cfg.Audit.MaxSizeMB)
	}
	if len(cfg.Rules.Custom) != 2 {
		t.Errorf("Custom rules: got %d, want 2", len(cfg.Rules.Custom))
	}
	if cfg.Rules.Custom[0].Name != "EMPLOYEE_ID" {
		t.Errorf("First custom rule: got %q", cfg.Rules.Custom[0].Name)
	}
	if len(cfg.Rules.Disabled) != 2 {
		t.Errorf("Disabled rules: got %d, want 2", len(cfg.Rules.Disabled))
	}

	t.Logf("Load valid config: PASSED (%d custom rules, %d disabled)", len(cfg.Rules.Custom), len(cfg.Rules.Disabled))
}
