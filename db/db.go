package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/David2024patton/itak-shield/auth"

	_ "modernc.org/sqlite"
)

// DB wraps the SQLite connection and provides all persistence operations.
type DB struct {
	mu   sync.RWMutex
	conn *sql.DB
	path string
}

// Open creates or opens a SQLite database at the given path.
// Tables are created automatically if they don't exist.
// If a legacy JSON file exists at jsonPath, data is migrated on first boot.
func Open(dbPath, legacyJSONPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, fmt.Errorf("db: cannot create directory %q: %w", dir, err)
		}
	}

	conn, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("db: cannot open %q: %w", dbPath, err)
	}

	// Enable WAL mode and foreign keys.
	if _, err := conn.Exec("PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("db: pragma setup failed: %w", err)
	}

	d := &DB{conn: conn, path: dbPath}

	if err := d.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("db: migration failed: %w", err)
	}

	// Migrate from legacy JSON if DB is empty and JSON exists.
	if legacyJSONPath != "" {
		if err := d.importLegacyJSON(legacyJSONPath); err != nil {
			// Non-fatal: log and continue.
			fmt.Printf("[shield] warning: legacy JSON import failed: %v\n", err)
		}
	}

	return d, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	if d == nil || d.conn == nil {
		return nil
	}
	return d.conn.Close()
}

// migrate creates all tables if they don't exist.
func (d *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id          TEXT PRIMARY KEY,
		name        TEXT NOT NULL,
		email       TEXT DEFAULT '',
		user_group  TEXT DEFAULT 'default',
		user_type   TEXT DEFAULT 'personal'  CHECK(user_type IN ('personal','business')),
		rate_limit  INTEGER DEFAULT 0,
		provider    TEXT DEFAULT '',
		upstream_key TEXT DEFAULT '',
		created_at  TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS tokens (
		key         TEXT PRIMARY KEY,
		user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		label       TEXT DEFAULT '',
		created_at  TEXT NOT NULL,
		expires_at  TEXT,
		revoked     INTEGER DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS spend (
		user_id       TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
		input_tokens  INTEGER DEFAULT 0,
		output_tokens INTEGER DEFAULT 0,
		estimated_usd REAL DEFAULT 0.0
	);

	CREATE TABLE IF NOT EXISTS pricing_profiles (
		id        TEXT PRIMARY KEY,
		name      TEXT NOT NULL,
		provider  TEXT NOT NULL DEFAULT '',
		input_per_million  REAL NOT NULL DEFAULT 3.00,
		output_per_million REAL NOT NULL DEFAULT 15.00,
		is_default INTEGER DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS user_billing (
		user_id       TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
		billing_type  TEXT DEFAULT 'pay_per_use' CHECK(billing_type IN ('pay_per_use','flat_rate')),
		pricing_id    TEXT DEFAULT '',
		monthly_budget REAL DEFAULT 0.0,
		period_start  TEXT DEFAULT '',
		period_end    TEXT DEFAULT '',
		notes         TEXT DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS audit_events (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp  TEXT NOT NULL,
		event_type TEXT NOT NULL,
		user_id    TEXT DEFAULT '',
		severity   TEXT DEFAULT '',
		details    TEXT DEFAULT '',
		source     TEXT DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS alert_configs (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		enabled         INTEGER DEFAULT 1,
		smtp_host       TEXT DEFAULT '',
		smtp_port       INTEGER DEFAULT 587,
		smtp_user       TEXT DEFAULT '',
		smtp_pass       TEXT DEFAULT '',
		smtp_tls        INTEGER DEFAULT 1,
		from_addr       TEXT DEFAULT '',
		recipients      TEXT DEFAULT '[]',
		min_severity    TEXT DEFAULT 'HIGH',
		digest_mode     TEXT DEFAULT 'instant' CHECK(digest_mode IN ('instant','hourly','daily')),
		updated_at      TEXT
	);

	CREATE TABLE IF NOT EXISTS settings (
		key   TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_tokens_user ON tokens(user_id);
	CREATE INDEX IF NOT EXISTS idx_audit_ts ON audit_events(timestamp);
	CREATE INDEX IF NOT EXISTS idx_audit_type ON audit_events(event_type);
	CREATE INDEX IF NOT EXISTS idx_audit_severity ON audit_events(severity);
	`
	_, err := d.conn.Exec(schema)
	return err
}

// importLegacyJSON migrates data from the old shield-users.json file.
// Only imports if the users table is empty.
func (d *DB) importLegacyJSON(jsonPath string) error {
	// Check if we already have users.
	var count int
	if err := d.conn.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil // Already have data, skip import.
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No legacy file, nothing to import.
		}
		return err
	}

	var users []auth.User
	if err := json.Unmarshal(data, &users); err != nil {
		return fmt.Errorf("parse legacy JSON: %w", err)
	}

	if len(users) == 0 {
		return nil
	}

	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, u := range users {
		_, err := tx.Exec(
			`INSERT OR IGNORE INTO users (id, name, email, user_group, rate_limit, created_at)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			u.ID, u.Name, u.Email, u.Group, u.RateLimit, u.CreatedAt.Format(time.RFC3339),
		)
		if err != nil {
			return fmt.Errorf("insert user %q: %w", u.Name, err)
		}

		for _, t := range u.Tokens {
			var expiresAt *string
			if t.ExpiresAt != nil {
				s := t.ExpiresAt.Format(time.RFC3339)
				expiresAt = &s
			}
			revoked := 0
			if t.Revoked {
				revoked = 1
			}
			_, err := tx.Exec(
				`INSERT OR IGNORE INTO tokens (key, user_id, label, created_at, expires_at, revoked)
				 VALUES (?, ?, ?, ?, ?, ?)`,
				t.Key, u.ID, t.Label, t.CreatedAt.Format(time.RFC3339), expiresAt, revoked,
			)
			if err != nil {
				return fmt.Errorf("insert token for %q: %w", u.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Printf("[shield] migrated %d users from legacy JSON -> SQLite\n", len(users))

	// Rename the old file so it doesn't get re-imported.
	_ = os.Rename(jsonPath, jsonPath+".migrated")

	return nil
}

// ─── auth.Store Interface ───────────────────────────

// Load returns all users with their tokens (implements auth.Store).
func (d *DB) Load() ([]auth.User, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rows, err := d.conn.Query(
		`SELECT id, name, email, user_group, user_type, rate_limit, provider, upstream_key, created_at
		 FROM users ORDER BY created_at`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []auth.User
	for rows.Next() {
		var u auth.User
		var createdStr, userType, provider, upstreamKey string
		err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Group, &userType, &u.RateLimit, &provider, &upstreamKey, &createdStr)
		if err != nil {
			return nil, err
		}
		u.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)

		// Load tokens for this user.
		tRows, err := d.conn.Query(
			`SELECT key, label, created_at, expires_at, revoked FROM tokens WHERE user_id = ?`, u.ID,
		)
		if err != nil {
			return nil, err
		}
		for tRows.Next() {
			var t auth.Token
			var created, expires sql.NullString
			var revoked int
			if err := tRows.Scan(&t.Key, &t.Label, &created, &expires, &revoked); err != nil {
				tRows.Close()
				return nil, err
			}
			t.CreatedAt, _ = time.Parse(time.RFC3339, created.String)
			if expires.Valid && expires.String != "" {
				exp, _ := time.Parse(time.RFC3339, expires.String)
				t.ExpiresAt = &exp
			}
			t.Revoked = revoked != 0
			u.Tokens = append(u.Tokens, t)
		}
		tRows.Close()

		users = append(users, u)
	}

	return users, rows.Err()
}

// Save persists all users and their tokens (implements auth.Store).
// Uses a transaction for atomicity.
func (d *DB) Save(users []auth.User) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, u := range users {
		_, err := tx.Exec(
			`INSERT INTO users (id, name, email, user_group, rate_limit, created_at)
			 VALUES (?, ?, ?, ?, ?, ?)
			 ON CONFLICT(id) DO UPDATE SET
			   name = excluded.name,
			   email = excluded.email,
			   user_group = excluded.user_group,
			   rate_limit = excluded.rate_limit`,
			u.ID, u.Name, u.Email, u.Group, u.RateLimit, u.CreatedAt.Format(time.RFC3339),
		)
		if err != nil {
			return err
		}

		for _, t := range u.Tokens {
			var expiresAt *string
			if t.ExpiresAt != nil {
				s := t.ExpiresAt.Format(time.RFC3339)
				expiresAt = &s
			}
			revoked := 0
			if t.Revoked {
				revoked = 1
			}
			_, err := tx.Exec(
				`INSERT INTO tokens (key, user_id, label, created_at, expires_at, revoked)
				 VALUES (?, ?, ?, ?, ?, ?)
				 ON CONFLICT(key) DO UPDATE SET
				   label = excluded.label,
				   expires_at = excluded.expires_at,
				   revoked = excluded.revoked`,
				t.Key, u.ID, t.Label, t.CreatedAt.Format(time.RFC3339), expiresAt, revoked,
			)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// ─── Extended User Profile Methods ──────────────────

// UserProfile is the enriched user data for the dashboard.
type UserProfile struct {
	auth.User
	Type        string  `json:"type"`         // "personal" or "business"
	Provider    string  `json:"provider"`     // "openai", "anthropic", etc.
	UpstreamKey string  `json:"upstream_key"` // their real API key (masked in responses)
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	EstimatedUSD float64 `json:"estimated_usd"`
	ActiveTokens int     `json:"active_tokens"`
	TotalTokens  int     `json:"total_tokens"`
}

// GetUserProfile returns enriched profile data for a single user.
func (d *DB) GetUserProfile(userID string) (*UserProfile, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	p := &UserProfile{}
	var createdStr string
	err := d.conn.QueryRow(
		`SELECT id, name, email, user_group, user_type, rate_limit, provider, upstream_key, created_at
		 FROM users WHERE id = ?`, userID,
	).Scan(&p.ID, &p.Name, &p.Email, &p.Group, &p.Type, &p.RateLimit, &p.Provider, &p.UpstreamKey, &createdStr)
	if err != nil {
		return nil, fmt.Errorf("user %q not found", userID)
	}
	p.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)

	// Get spend.
	_ = d.conn.QueryRow(
		`SELECT COALESCE(input_tokens,0), COALESCE(output_tokens,0), COALESCE(estimated_usd,0)
		 FROM spend WHERE user_id = ?`, userID,
	).Scan(&p.InputTokens, &p.OutputTokens, &p.EstimatedUSD)

	// Get token counts.
	_ = d.conn.QueryRow(
		`SELECT COUNT(*), SUM(CASE WHEN revoked=0 THEN 1 ELSE 0 END)
		 FROM tokens WHERE user_id = ?`, userID,
	).Scan(&p.TotalTokens, &p.ActiveTokens)

	return p, nil
}

// ListUserProfiles returns enriched profiles for all users.
func (d *DB) ListUserProfiles() ([]UserProfile, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rows, err := d.conn.Query(
		`SELECT u.id, u.name, u.email, u.user_group, u.user_type, u.rate_limit,
		        u.provider, u.upstream_key, u.created_at,
		        COALESCE(s.input_tokens, 0), COALESCE(s.output_tokens, 0),
		        COALESCE(s.estimated_usd, 0)
		 FROM users u
		 LEFT JOIN spend s ON s.user_id = u.id
		 ORDER BY u.created_at`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []UserProfile
	for rows.Next() {
		var p UserProfile
		var createdStr string
		err := rows.Scan(
			&p.ID, &p.Name, &p.Email, &p.Group, &p.Type, &p.RateLimit,
			&p.Provider, &p.UpstreamKey, &createdStr,
			&p.InputTokens, &p.OutputTokens, &p.EstimatedUSD,
		)
		if err != nil {
			return nil, err
		}
		p.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		profiles = append(profiles, p)
	}

	// Get token counts for each user.
	for i := range profiles {
		_ = d.conn.QueryRow(
			`SELECT COUNT(*), SUM(CASE WHEN revoked=0 THEN 1 ELSE 0 END)
			 FROM tokens WHERE user_id = ?`, profiles[i].ID,
		).Scan(&profiles[i].TotalTokens, &profiles[i].ActiveTokens)
	}

	return profiles, rows.Err()
}

// UpdateUserProfile updates the extended profile fields.
func (d *DB) UpdateUserProfile(userID, userType, provider, upstreamKey string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	result, err := d.conn.Exec(
		`UPDATE users SET user_type = ?, provider = ?, upstream_key = ? WHERE id = ?`,
		userType, provider, upstreamKey, userID,
	)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user %q not found", userID)
	}
	return nil
}

// GetUserUpstreamKey returns the real upstream API key for a user.
// Used by the proxy to route requests through the user's own provider.
func (d *DB) GetUserUpstreamKey(userID string) (provider, upstreamKey string, err error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	err = d.conn.QueryRow(
		`SELECT provider, upstream_key FROM users WHERE id = ?`, userID,
	).Scan(&provider, &upstreamKey)
	return
}

// ─── Spend Tracking ─────────────────────────────────

// RecordSpend adds token usage for a user.
func (d *DB) RecordSpend(userID string, inputTokens, outputTokens int64, costUSD float64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.conn.Exec(
		`INSERT INTO spend (user_id, input_tokens, output_tokens, estimated_usd)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(user_id) DO UPDATE SET
		   input_tokens = input_tokens + excluded.input_tokens,
		   output_tokens = output_tokens + excluded.output_tokens,
		   estimated_usd = estimated_usd + excluded.estimated_usd`,
		userID, inputTokens, outputTokens, costUSD,
	)
	return err
}

// ─── Audit Events ───────────────────────────────────

// LogEvent records a security or system event.
func (d *DB) LogEvent(eventType, userID, severity, details, source string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.conn.Exec(
		`INSERT INTO audit_events (timestamp, event_type, user_id, severity, details, source)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		time.Now().UTC().Format(time.RFC3339), eventType, userID, severity, details, source,
	)
	return err
}

// RecentEvents returns the last N audit events, optionally filtered by type.
func (d *DB) RecentEvents(limit int, eventType string) ([]map[string]interface{}, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var rows *sql.Rows
	var err error
	if eventType != "" {
		rows, err = d.conn.Query(
			`SELECT id, timestamp, event_type, user_id, severity, details, source
			 FROM audit_events WHERE event_type = ? ORDER BY id DESC LIMIT ?`,
			eventType, limit,
		)
	} else {
		rows, err = d.conn.Query(
			`SELECT id, timestamp, event_type, user_id, severity, details, source
			 FROM audit_events ORDER BY id DESC LIMIT ?`,
			limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var id int64
		var ts, evType, uid, sev, details, src string
		if err := rows.Scan(&id, &ts, &evType, &uid, &sev, &details, &src); err != nil {
			return nil, err
		}
		events = append(events, map[string]interface{}{
			"id":         id,
			"timestamp":  ts,
			"event_type": evType,
			"user_id":    uid,
			"severity":   sev,
			"details":    details,
			"source":     src,
		})
	}
	return events, rows.Err()
}

// ─── Settings KV ────────────────────────────────────

// GetSetting returns a setting value by key.
func (d *DB) GetSetting(key string) (string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var val string
	err := d.conn.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&val)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return val, err
}

// SetSetting saves a key-value setting.
func (d *DB) SetSetting(key, value string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.conn.Exec(
		`INSERT INTO settings (key, value) VALUES (?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value,
	)
	return err
}

// ─── Pricing Profiles ───────────────────────────────

// PricingProfile defines token costs for a specific provider/model tier.
type PricingProfile struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Provider         string  `json:"provider"`
	InputPerMillion  float64 `json:"input_per_million"`
	OutputPerMillion float64 `json:"output_per_million"`
	IsDefault        bool    `json:"is_default"`
}

// ListPricingProfiles returns all pricing profiles.
func (d *DB) ListPricingProfiles() ([]PricingProfile, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rows, err := d.conn.Query(
		`SELECT id, name, provider, input_per_million, output_per_million, is_default
		 FROM pricing_profiles ORDER BY provider, name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []PricingProfile
	for rows.Next() {
		var p PricingProfile
		var isDefault int
		if err := rows.Scan(&p.ID, &p.Name, &p.Provider, &p.InputPerMillion, &p.OutputPerMillion, &isDefault); err != nil {
			return nil, err
		}
		p.IsDefault = isDefault != 0
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}

// UpsertPricingProfile creates or updates a pricing profile.
func (d *DB) UpsertPricingProfile(p PricingProfile) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	isDefault := 0
	if p.IsDefault {
		isDefault = 1
	}

	_, err := d.conn.Exec(
		`INSERT INTO pricing_profiles (id, name, provider, input_per_million, output_per_million, is_default)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   name = excluded.name,
		   provider = excluded.provider,
		   input_per_million = excluded.input_per_million,
		   output_per_million = excluded.output_per_million,
		   is_default = excluded.is_default`,
		p.ID, p.Name, p.Provider, p.InputPerMillion, p.OutputPerMillion, isDefault,
	)
	return err
}

// DeletePricingProfile removes a pricing profile.
func (d *DB) DeletePricingProfile(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.conn.Exec(`DELETE FROM pricing_profiles WHERE id = ?`, id)
	return err
}

// SeedDefaultPricing inserts common provider pricing if none exist.
func (d *DB) SeedDefaultPricing() error {
	var count int
	d.conn.QueryRow(`SELECT COUNT(*) FROM pricing_profiles`).Scan(&count)
	if count > 0 {
		return nil // Already have pricing data.
	}

	defaults := []PricingProfile{
		// ─── OpenAI (openai.com/pricing, March 2026) ───
		{ID: "openai-gpt4o", Name: "GPT-4o", Provider: "openai", InputPerMillion: 2.50, OutputPerMillion: 10.00, IsDefault: true},
		{ID: "openai-gpt4o-mini", Name: "GPT-4o Mini", Provider: "openai", InputPerMillion: 0.15, OutputPerMillion: 0.60},
		{ID: "openai-o1", Name: "o1", Provider: "openai", InputPerMillion: 15.00, OutputPerMillion: 60.00},
		{ID: "openai-o3", Name: "o3", Provider: "openai", InputPerMillion: 2.00, OutputPerMillion: 8.00},

		// ─── Anthropic (anthropic.com/pricing, March 2026) ───
		{ID: "anthropic-opus", Name: "Claude Opus 4.6", Provider: "anthropic", InputPerMillion: 5.00, OutputPerMillion: 25.00},
		{ID: "anthropic-sonnet", Name: "Claude Sonnet 4.6", Provider: "anthropic", InputPerMillion: 3.00, OutputPerMillion: 15.00, IsDefault: true},
		{ID: "anthropic-haiku", Name: "Claude Haiku 4.5", Provider: "anthropic", InputPerMillion: 1.00, OutputPerMillion: 5.00},

		// ─── Google (ai.google.dev/pricing, March 2026) ───
		{ID: "google-gemini3-pro", Name: "Gemini 3 Pro", Provider: "google", InputPerMillion: 2.00, OutputPerMillion: 12.00},
		{ID: "google-gemini3-flash", Name: "Gemini 3 Flash", Provider: "google", InputPerMillion: 0.50, OutputPerMillion: 3.00, IsDefault: true},
		{ID: "google-gemini25-pro", Name: "Gemini 2.5 Pro", Provider: "google", InputPerMillion: 1.25, OutputPerMillion: 10.00},
		{ID: "google-gemini25-flash", Name: "Gemini 2.5 Flash", Provider: "google", InputPerMillion: 0.30, OutputPerMillion: 2.50},

		// ─── xAI (x.ai/api, March 2026) ───
		{ID: "xai-grok4", Name: "Grok 4", Provider: "xai", InputPerMillion: 3.00, OutputPerMillion: 15.00, IsDefault: true},
		{ID: "xai-grok4-fast", Name: "Grok 4 Fast", Provider: "xai", InputPerMillion: 0.20, OutputPerMillion: 0.50},

		// ─── DeepSeek (deepseek.com/pricing, March 2026) ───
		{ID: "deepseek-v32", Name: "DeepSeek V3.2", Provider: "deepseek", InputPerMillion: 0.28, OutputPerMillion: 0.42, IsDefault: true},
		{ID: "deepseek-r1", Name: "DeepSeek R1", Provider: "deepseek", InputPerMillion: 0.55, OutputPerMillion: 2.19},

		// ─── Mistral (mistral.ai/pricing, March 2026) ───
		{ID: "mistral-large3", Name: "Mistral Large 3", Provider: "mistral", InputPerMillion: 0.50, OutputPerMillion: 1.50, IsDefault: true},
		{ID: "mistral-small31", Name: "Mistral Small 3.1", Provider: "mistral", InputPerMillion: 0.10, OutputPerMillion: 0.30},
		{ID: "mistral-codestral", Name: "Codestral", Provider: "mistral", InputPerMillion: 0.30, OutputPerMillion: 0.90},

		// ─── Cohere (cohere.com/pricing, March 2026) ───
		{ID: "cohere-command-r-plus", Name: "Command R+", Provider: "cohere", InputPerMillion: 2.50, OutputPerMillion: 10.00, IsDefault: true},
		{ID: "cohere-command-r", Name: "Command R", Provider: "cohere", InputPerMillion: 0.15, OutputPerMillion: 0.60},

		// ─── NVIDIA NIM (build.nvidia.com, March 2026) ───
		{ID: "nvidia-nemotron-nano", Name: "Nemotron Nano 30B", Provider: "nvidia", InputPerMillion: 0.05, OutputPerMillion: 0.20, IsDefault: true},

		// ─── Qwen / Alibaba (alibabacloud.com/dashscope, March 2026) ───
		{ID: "qwen-max", Name: "Qwen Max", Provider: "qwen", InputPerMillion: 1.60, OutputPerMillion: 6.40, IsDefault: true},
		{ID: "qwen-plus", Name: "Qwen Plus", Provider: "qwen", InputPerMillion: 0.40, OutputPerMillion: 1.20},
		{ID: "qwen-turbo", Name: "Qwen Turbo", Provider: "qwen", InputPerMillion: 0.05, OutputPerMillion: 0.20},
		{ID: "qwen-coder", Name: "Qwen Coder", Provider: "qwen", InputPerMillion: 0.30, OutputPerMillion: 1.50},

		// ─── Kimi / Moonshot (platform.moonshot.ai, March 2026) ───
		{ID: "kimi-v1-8k", Name: "Moonshot v1 8K", Provider: "kimi", InputPerMillion: 0.20, OutputPerMillion: 2.00, IsDefault: true},
		{ID: "kimi-v1-32k", Name: "Moonshot v1 32K", Provider: "kimi", InputPerMillion: 1.00, OutputPerMillion: 3.00},
		{ID: "kimi-v1-128k", Name: "Moonshot v1 128K", Provider: "kimi", InputPerMillion: 2.00, OutputPerMillion: 5.00},

		// ─── Zhipu AI / GLM (z.ai, March 2026) ───
		{ID: "zhipu-glm47", Name: "GLM-4.7", Provider: "zhipu", InputPerMillion: 0.60, OutputPerMillion: 2.20, IsDefault: true},
		{ID: "zhipu-glm46", Name: "GLM-4.6", Provider: "zhipu", InputPerMillion: 0.60, OutputPerMillion: 2.20},
		{ID: "zhipu-glm47-flashx", Name: "GLM-4.7 FlashX", Provider: "zhipu", InputPerMillion: 0.07, OutputPerMillion: 0.40},

		// ─── Meta AI (llama.com, March 2026) ───
		{ID: "meta-llama4", Name: "Llama 4 Maverick", Provider: "meta", InputPerMillion: 0.00, OutputPerMillion: 0.00, IsDefault: true},

		// ─── OpenRouter (openrouter.ai, pass-through pricing, March 2026) ───
		{ID: "openrouter-default", Name: "OpenRouter (varies by model)", Provider: "openrouter", InputPerMillion: 1.00, OutputPerMillion: 5.00, IsDefault: true},

		// ─── Groq (groq.com/pricing, March 2026) ───
		{ID: "groq-llama33-70b", Name: "Llama 3.3 70B", Provider: "groq", InputPerMillion: 0.59, OutputPerMillion: 0.79, IsDefault: true},
		{ID: "groq-llama31-8b", Name: "Llama 3.1 8B", Provider: "groq", InputPerMillion: 0.05, OutputPerMillion: 0.08},
		{ID: "groq-deepseek-r1", Name: "DeepSeek R1 Distill 70B", Provider: "groq", InputPerMillion: 0.75, OutputPerMillion: 0.99},

		// ─── Together AI (together.ai/pricing, March 2026) ───
		{ID: "together-llama33-70b", Name: "Llama 3.3 70B", Provider: "together", InputPerMillion: 0.88, OutputPerMillion: 0.88, IsDefault: true},
		{ID: "together-llama31-405b", Name: "Llama 3.1 405B", Provider: "together", InputPerMillion: 3.50, OutputPerMillion: 3.50},

		// ─── Fireworks AI (fireworks.ai/pricing, March 2026) ───
		{ID: "fireworks-llama33-70b", Name: "Llama 3.3 70B", Provider: "fireworks", InputPerMillion: 0.90, OutputPerMillion: 0.90, IsDefault: true},
		{ID: "fireworks-llama31-405b", Name: "Llama 3.1 405B", Provider: "fireworks", InputPerMillion: 3.00, OutputPerMillion: 3.00},

		// ─── Hugging Face (huggingface.co, compute-based, March 2026) ───
		{ID: "hf-serverless", Name: "HF Serverless (estimate)", Provider: "huggingface", InputPerMillion: 0.10, OutputPerMillion: 0.30, IsDefault: true},

		// ─── MiniMax (minimaxi.com, March 2026) ───
		{ID: "minimax-m25", Name: "MiniMax M2.5", Provider: "minimax", InputPerMillion: 0.27, OutputPerMillion: 0.95, IsDefault: true},

		// ─── Perplexity (perplexity.ai/pricing, March 2026) ───
		{ID: "perplexity-sonar", Name: "Sonar", Provider: "perplexity", InputPerMillion: 1.00, OutputPerMillion: 1.00, IsDefault: true},
		{ID: "perplexity-sonar-pro", Name: "Sonar Pro", Provider: "perplexity", InputPerMillion: 3.00, OutputPerMillion: 15.00},

		// ─── Cerebras (cerebras.ai/pricing, March 2026) ───
		{ID: "cerebras-llama31-70b", Name: "Llama 3.1 70B", Provider: "cerebras", InputPerMillion: 0.60, OutputPerMillion: 0.60, IsDefault: true},

		// ─── DeepInfra (deepinfra.com, March 2026) ───
		{ID: "deepinfra-llama33-70b", Name: "Llama 3.3 70B", Provider: "deepinfra", InputPerMillion: 0.23, OutputPerMillion: 0.40, IsDefault: true},

		// ─── SiliconFlow (siliconflow.cn, March 2026) ───
		{ID: "siliconflow-default", Name: "SiliconFlow (varies)", Provider: "siliconflow", InputPerMillion: 0.10, OutputPerMillion: 0.30, IsDefault: true},

		// ─── Manus AI (manus.im, agent platform, March 2026) ───
		{ID: "manus-agent", Name: "Manus Agent", Provider: "manus", InputPerMillion: 0.00, OutputPerMillion: 0.00, IsDefault: true},

		// ─── Custom (user enters their own rates) ───
		{ID: "custom-default", Name: "Custom Provider", Provider: "custom", InputPerMillion: 1.00, OutputPerMillion: 3.00, IsDefault: true},

		// ─── Local / Self-hosted (always free) ───
		{ID: "local-free", Name: "Local / Ollama", Provider: "local", InputPerMillion: 0.00, OutputPerMillion: 0.00, IsDefault: true},
	}

	for _, p := range defaults {
		if err := d.UpsertPricingProfile(p); err != nil {
			return err
		}
	}
	return nil
}

// ─── User Billing ───────────────────────────────────

// UserBillingInfo tracks a user's billing configuration.
type UserBillingInfo struct {
	UserID        string  `json:"user_id"`
	BillingType   string  `json:"billing_type"`   // "pay_per_use" or "flat_rate"
	PricingID     string  `json:"pricing_id"`     // references pricing_profiles.id
	MonthlyBudget float64 `json:"monthly_budget"` // max USD per period (0 = unlimited)
	PeriodStart   string  `json:"period_start"`
	PeriodEnd     string  `json:"period_end"`
	Notes         string  `json:"notes"`
}

// GetUserBilling returns billing info for a user.
func (d *DB) GetUserBilling(userID string) (*UserBillingInfo, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	b := &UserBillingInfo{UserID: userID, BillingType: "pay_per_use"}
	err := d.conn.QueryRow(
		`SELECT billing_type, pricing_id, monthly_budget, period_start, period_end, notes
		 FROM user_billing WHERE user_id = ?`, userID,
	).Scan(&b.BillingType, &b.PricingID, &b.MonthlyBudget, &b.PeriodStart, &b.PeriodEnd, &b.Notes)
	if err == sql.ErrNoRows {
		return b, nil // Return defaults.
	}
	return b, err
}

// UpsertUserBilling creates or updates billing info for a user.
func (d *DB) UpsertUserBilling(b UserBillingInfo) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.conn.Exec(
		`INSERT INTO user_billing (user_id, billing_type, pricing_id, monthly_budget, period_start, period_end, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(user_id) DO UPDATE SET
		   billing_type = excluded.billing_type,
		   pricing_id = excluded.pricing_id,
		   monthly_budget = excluded.monthly_budget,
		   period_start = excluded.period_start,
		   period_end = excluded.period_end,
		   notes = excluded.notes`,
		b.UserID, b.BillingType, b.PricingID, b.MonthlyBudget, b.PeriodStart, b.PeriodEnd, b.Notes,
	)
	return err
}

// GetEffectivePricing resolves the pricing for a user: user billing -> provider default -> system default.
func (d *DB) GetEffectivePricing(userID string) (inputPerMillion, outputPerMillion float64, err error) {
	// Default system pricing.
	inputPerMillion = 3.00
	outputPerMillion = 15.00

	d.mu.RLock()
	defer d.mu.RUnlock()

	// Check user billing for a specific pricing profile.
	var pricingID string
	err = d.conn.QueryRow(
		`SELECT pricing_id FROM user_billing WHERE user_id = ?`, userID,
	).Scan(&pricingID)
	if err == nil && pricingID != "" {
		// Use their specific pricing profile.
		err = d.conn.QueryRow(
			`SELECT input_per_million, output_per_million FROM pricing_profiles WHERE id = ?`, pricingID,
		).Scan(&inputPerMillion, &outputPerMillion)
		if err == nil {
			return inputPerMillion, outputPerMillion, nil
		}
	}

	// Fall back to provider default pricing.
	var provider string
	_ = d.conn.QueryRow(`SELECT provider FROM users WHERE id = ?`, userID).Scan(&provider)
	if provider != "" {
		err = d.conn.QueryRow(
			`SELECT input_per_million, output_per_million FROM pricing_profiles
			 WHERE provider = ? AND is_default = 1 LIMIT 1`, provider,
		).Scan(&inputPerMillion, &outputPerMillion)
		if err == nil {
			return inputPerMillion, outputPerMillion, nil
		}
	}

	return inputPerMillion, outputPerMillion, nil
}

// ResetUserSpend clears accumulated spend for a user (e.g. at billing period reset).
func (d *DB) ResetUserSpend(userID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.conn.Exec(
		`UPDATE spend SET input_tokens = 0, output_tokens = 0, estimated_usd = 0 WHERE user_id = ?`,
		userID,
	)
	return err
}
