package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/David2024patton/itak-shield/auth"
)

func tempDB(t *testing.T) *DB {
	t.Helper()
	dir := t.TempDir()
	d, err := Open(filepath.Join(dir, "test.db"), "")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func TestOpenAndMigrate(t *testing.T) {
	d := tempDB(t)
	if d == nil {
		t.Fatal("db is nil")
	}
}

func TestLoadSaveRoundTrip(t *testing.T) {
	d := tempDB(t)

	now := time.Now().Truncate(time.Second)
	exp := now.Add(24 * time.Hour)

	users := []auth.User{
		{
			ID: "u1", Name: "Alice", Email: "alice@test.com",
			Group: "devs", RateLimit: 60, CreatedAt: now,
			Tokens: []auth.Token{
				{Key: "shield_abc123", Label: "laptop", CreatedAt: now, ExpiresAt: &exp},
				{Key: "shield_def456", Label: "revoked", CreatedAt: now, Revoked: true},
			},
		},
		{
			ID: "u2", Name: "Bob", Email: "",
			Group: "default", RateLimit: 0, CreatedAt: now,
			Tokens: []auth.Token{},
		},
	}

	if err := d.Save(users); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := d.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("expected 2 users, got %d", len(loaded))
	}

	// Find Alice.
	var alice auth.User
	for _, u := range loaded {
		if u.ID == "u1" {
			alice = u
		}
	}
	if alice.Name != "Alice" {
		t.Errorf("expected Alice, got %q", alice.Name)
	}
	if len(alice.Tokens) != 2 {
		t.Errorf("expected 2 tokens, got %d", len(alice.Tokens))
	}
}

func TestLegacyJSONMigration(t *testing.T) {
	dir := t.TempDir()

	// Create a legacy JSON file.
	now := time.Now().Truncate(time.Second)

	jsonPath := filepath.Join(dir, "shield-users.json")

	// Write legacy JSON.
	importData := `[{"id":"legacy1","name":"LegacyUser","email":"old@test.com","group":"default","rate_limit":30,"created_at":"` + now.Format(time.RFC3339) + `","tokens":[{"key":"shield_old_key","label":"old","created_at":"` + now.Format(time.RFC3339) + `"}]}]`
	if err := os.WriteFile(jsonPath, []byte(importData), 0644); err != nil {
		t.Fatalf("write json: %v", err)
	}

	d, err := Open(filepath.Join(dir, "test.db"), jsonPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer d.Close()

	loaded, err := d.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("expected 1 migrated user, got %d", len(loaded))
	}
	if loaded[0].Name != "LegacyUser" {
		t.Errorf("expected LegacyUser, got %q", loaded[0].Name)
	}
	if len(loaded[0].Tokens) != 1 {
		t.Errorf("expected 1 token, got %d", len(loaded[0].Tokens))
	}

	// JSON file should be renamed.
	if _, err := os.Stat(jsonPath + ".migrated"); err != nil {
		t.Errorf("expected .migrated file: %v", err)
	}
}

func TestUserProfile(t *testing.T) {
	d := tempDB(t)

	users := []auth.User{
		{
			ID: "p1", Name: "ProfileUser", Email: "profile@test.com",
			Group: "team", RateLimit: 100, CreatedAt: time.Now(),
			Tokens: []auth.Token{
				{Key: "shield_tk1", Label: "main", CreatedAt: time.Now()},
				{Key: "shield_tk2", Label: "backup", CreatedAt: time.Now(), Revoked: true},
			},
		},
	}
	if err := d.Save(users); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Set profile fields.
	if err := d.UpdateUserProfile("p1", "business", "openai", "sk-real-key-123"); err != nil {
		t.Fatalf("update: %v", err)
	}

	// Record some spend.
	if err := d.RecordSpend("p1", 5000, 1500, 0.023); err != nil {
		t.Fatalf("record spend: %v", err)
	}

	profile, err := d.GetUserProfile("p1")
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}

	if profile.Type != "business" {
		t.Errorf("type: got %q, want 'business'", profile.Type)
	}
	if profile.Provider != "openai" {
		t.Errorf("provider: got %q, want 'openai'", profile.Provider)
	}
	if profile.UpstreamKey != "sk-real-key-123" {
		t.Errorf("upstream key mismatch")
	}
	if profile.InputTokens != 5000 {
		t.Errorf("input tokens: got %d, want 5000", profile.InputTokens)
	}
	if profile.ActiveTokens != 1 {
		t.Errorf("active tokens: got %d, want 1", profile.ActiveTokens)
	}
	if profile.TotalTokens != 2 {
		t.Errorf("total tokens: got %d, want 2", profile.TotalTokens)
	}
}

func TestAuditEvents(t *testing.T) {
	d := tempDB(t)

	if err := d.LogEvent("guard_block", "u1", "CRITICAL", "instruction_override detected", "email"); err != nil {
		t.Fatalf("log event: %v", err)
	}
	if err := d.LogEvent("pii_redacted", "u1", "INFO", "SSN found and redacted", "user"); err != nil {
		t.Fatalf("log event: %v", err)
	}

	events, err := d.RecentEvents(10, "")
	if err != nil {
		t.Fatalf("recent events: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}

	// Filter by type.
	filtered, err := d.RecentEvents(10, "guard_block")
	if err != nil {
		t.Fatalf("filtered events: %v", err)
	}
	if len(filtered) != 1 {
		t.Errorf("expected 1 guard_block, got %d", len(filtered))
	}
}

func TestSettings(t *testing.T) {
	d := tempDB(t)

	if err := d.SetSetting("auto_start", "true"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := d.SetSetting("smtp_host", "smtp.gmail.com"); err != nil {
		t.Fatalf("set: %v", err)
	}

	val, err := d.GetSetting("auto_start")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if val != "true" {
		t.Errorf("got %q, want 'true'", val)
	}

	// Overwrite.
	if err := d.SetSetting("auto_start", "false"); err != nil {
		t.Fatalf("set: %v", err)
	}
	val, _ = d.GetSetting("auto_start")
	if val != "false" {
		t.Errorf("got %q, want 'false'", val)
	}

	// Non-existent key returns empty.
	val, _ = d.GetSetting("nonexistent")
	if val != "" {
		t.Errorf("got %q, want empty", val)
	}
}

func TestListUserProfiles(t *testing.T) {
	d := tempDB(t)

	users := []auth.User{
		{ID: "a1", Name: "Alpha", Group: "team", CreatedAt: time.Now(), Tokens: []auth.Token{{Key: "k1", Label: "x", CreatedAt: time.Now()}}},
		{ID: "a2", Name: "Beta", Group: "team", CreatedAt: time.Now(), Tokens: []auth.Token{}},
	}
	d.Save(users)
	d.UpdateUserProfile("a1", "business", "anthropic", "sk-ant-xxx")
	d.UpdateUserProfile("a2", "personal", "openai", "sk-oai-yyy")
	d.RecordSpend("a1", 10000, 3000, 0.05)

	profiles, err := d.ListUserProfiles()
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}

	// Check Alpha specifically.
	for _, p := range profiles {
		if p.Name == "Alpha" {
			if p.Provider != "anthropic" {
				t.Errorf("provider: got %q", p.Provider)
			}
			if p.InputTokens != 10000 {
				t.Errorf("input: got %d", p.InputTokens)
			}
		}
	}
}
