package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ── Data Structures ─────────────────────────────

// User represents a team member who can have multiple API tokens.
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email,omitempty"`
	Group     string    `json:"group"`
	RateLimit int       `json:"rate_limit"` // requests per minute, 0 = unlimited
	CreatedAt time.Time `json:"created_at"`
	Tokens    []Token   `json:"tokens"`
}

// Token is an API key bound to a user with optional expiration.
type Token struct {
	Key       string     `json:"key"`
	Label     string     `json:"label"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"` // nil = never expires
	Revoked   bool       `json:"revoked"`
}

// Identity is the authenticated user context attached to each request.
type Identity struct {
	UserID string
	User   string
	Group  string
}

// Stats holds authentication statistics for the analytics dashboard.
type Stats struct {
	RequestsByUser map[string]int64 `json:"requests_by_user"`
	BlockedByRate  int64            `json:"blocked_by_rate"`
}

// KeyEntry is kept for backward compatibility with YAML config-based keys.
type KeyEntry struct {
	Key       string `yaml:"key"`
	User      string `yaml:"user"`
	Group     string `yaml:"group"`
	RateLimit int    `yaml:"rate_limit"`
}

// ── Manager ─────────────────────────────────────

// Manager handles user accounts, API token authentication, and rate limiting.
type Manager struct {
	mu        sync.RWMutex
	users     map[string]*User       // user ID -> User
	tokenIdx  map[string]string      // token key -> user ID (lookup index)
	windows   map[string]*rateWindow // user ID -> sliding window
	injectKey string                 // real upstream API key to inject
	store     Store                  // persistence backend (can be nil)
}

// rateWindow tracks request counts in a sliding window per user.
type rateWindow struct {
	mu      sync.Mutex
	limit   int
	entries []time.Time
}

// New creates a Manager, optionally loading persisted state from a Store.
// If store is nil, tokens exist only in memory.
func New(store Store, injectKey string) *Manager {
	m := &Manager{
		users:     make(map[string]*User),
		tokenIdx:  make(map[string]string),
		windows:   make(map[string]*rateWindow),
		injectKey: injectKey,
		store:     store,
	}

	// Load persisted users if a store is provided.
	if store != nil {
		users, err := store.Load()
		if err == nil {
			for _, u := range users {
				uc := u // copy
				m.users[uc.ID] = &uc
				for _, t := range uc.Tokens {
					if !t.Revoked {
						m.tokenIdx[t.Key] = uc.ID
					}
				}
				if uc.RateLimit > 0 {
					m.windows[uc.ID] = &rateWindow{
						limit:   uc.RateLimit,
						entries: make([]time.Time, 0, uc.RateLimit),
					}
				}
			}
		}
	}

	return m
}

// NewFromEntries creates a Manager from legacy KeyEntry config (backward compat).
func NewFromEntries(entries []KeyEntry, injectKey string) *Manager {
	m := New(nil, injectKey)
	for _, e := range entries {
		userID := generateID()
		u := &User{
			ID:        userID,
			Name:      e.User,
			Group:     e.Group,
			RateLimit: e.RateLimit,
			CreatedAt: time.Now(),
			Tokens: []Token{{
				Key:       e.Key,
				Label:     "config",
				CreatedAt: time.Now(),
			}},
		}
		m.users[userID] = u
		m.tokenIdx[e.Key] = userID
		if e.RateLimit > 0 {
			m.windows[userID] = &rateWindow{
				limit:   e.RateLimit,
				entries: make([]time.Time, 0, e.RateLimit),
			}
		}
	}
	return m
}

// ── User CRUD ───────────────────────────────────

// CreateUser adds a new user and returns their ID.
func (m *Manager) CreateUser(name, email, group string, rateLimit int) (*User, error) {
	if m == nil {
		return nil, fmt.Errorf("auth manager not initialized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	u := &User{
		ID:        generateID(),
		Name:      name,
		Email:     email,
		Group:     group,
		RateLimit: rateLimit,
		CreatedAt: time.Now(),
		Tokens:    []Token{},
	}
	m.users[u.ID] = u

	if rateLimit > 0 {
		m.windows[u.ID] = &rateWindow{
			limit:   rateLimit,
			entries: make([]time.Time, 0, rateLimit),
		}
	}

	m.persist()
	return u, nil
}

// DeleteUser removes a user and revokes all their tokens.
func (m *Manager) DeleteUser(userID string) error {
	if m == nil {
		return fmt.Errorf("auth manager not initialized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	u, ok := m.users[userID]
	if !ok {
		return fmt.Errorf("user not found: %s", userID)
	}

	// Remove all token index entries.
	for _, t := range u.Tokens {
		delete(m.tokenIdx, t.Key)
	}
	delete(m.users, userID)
	delete(m.windows, userID)

	m.persist()
	return nil
}

// ListUsers returns all users (shallow copy for safety).
func (m *Manager) ListUsers() []User {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]User, 0, len(m.users))
	for _, u := range m.users {
		result = append(result, *u)
	}
	return result
}

// GetUser returns a single user by ID.
func (m *Manager) GetUser(userID string) (*User, error) {
	if m == nil {
		return nil, fmt.Errorf("auth manager not initialized")
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	u, ok := m.users[userID]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", userID)
	}
	copy := *u
	return &copy, nil
}

// ── Token CRUD ──────────────────────────────────

// GenerateToken creates a new API token for a user.
// If expiresAt is nil, the token never expires.
func (m *Manager) GenerateToken(userID, label string, expiresAt *time.Time) (*Token, error) {
	if m == nil {
		return nil, fmt.Errorf("auth manager not initialized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	u, ok := m.users[userID]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", userID)
	}

	token := Token{
		Key:       generateToken(),
		Label:     label,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}
	u.Tokens = append(u.Tokens, token)
	m.tokenIdx[token.Key] = userID

	m.persist()
	return &token, nil
}

// RevokeToken marks a token as revoked and removes it from the lookup index.
func (m *Manager) RevokeToken(userID, tokenKey string) error {
	if m == nil {
		return fmt.Errorf("auth manager not initialized")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	u, ok := m.users[userID]
	if !ok {
		return fmt.Errorf("user not found: %s", userID)
	}

	found := false
	for i := range u.Tokens {
		if u.Tokens[i].Key == tokenKey {
			u.Tokens[i].Revoked = true
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("token not found")
	}

	delete(m.tokenIdx, tokenKey)
	m.persist()
	return nil
}

// ── Authentication ──────────────────────────────

// Authenticate validates the Authorization header and returns the user identity.
func (m *Manager) Authenticate(authHeader string) (*Identity, error) {
	if m == nil {
		return &Identity{User: "anonymous", Group: "default"}, nil
	}

	m.mu.RLock()
	userCount := len(m.users)
	m.mu.RUnlock()

	// No users registered = auth disabled, pass everything through.
	if userCount == 0 {
		return &Identity{User: "anonymous", Group: "default"}, nil
	}

	// Extract bearer token.
	token := ""
	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	} else if strings.HasPrefix(authHeader, "bearer ") {
		token = strings.TrimPrefix(authHeader, "bearer ")
	} else {
		return nil, fmt.Errorf("missing or invalid Authorization header")
	}

	m.mu.RLock()
	userID, ok := m.tokenIdx[token]
	if !ok {
		m.mu.RUnlock()
		return nil, fmt.Errorf("invalid API key")
	}
	u := m.users[userID]
	m.mu.RUnlock()

	// Check if the specific token is expired.
	for _, t := range u.Tokens {
		if t.Key == token {
			if t.ExpiresAt != nil && time.Now().After(*t.ExpiresAt) {
				return nil, fmt.Errorf("token expired")
			}
			break
		}
	}

	return &Identity{UserID: u.ID, User: u.Name, Group: u.Group}, nil
}

// CheckRateLimit enforces the per-user request rate limit.
func (m *Manager) CheckRateLimit(user string) error {
	if m == nil {
		return nil
	}

	m.mu.RLock()
	w, ok := m.windows[user]
	m.mu.RUnlock()

	if !ok {
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Minute)

	valid := 0
	for _, t := range w.entries {
		if t.After(cutoff) {
			w.entries[valid] = t
			valid++
		}
	}
	w.entries = w.entries[:valid]

	if len(w.entries) >= w.limit {
		return fmt.Errorf("rate limit exceeded: %d requests/min for user %s", w.limit, user)
	}

	w.entries = append(w.entries, now)
	return nil
}

// InjectKey returns the real upstream API key to inject into requests.
func (m *Manager) InjectKey() string {
	if m == nil {
		return ""
	}
	return m.injectKey
}

// HasKeys returns true if any users with active tokens are configured.
func (m *Manager) HasKeys() bool {
	if m == nil {
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.tokenIdx) > 0
}

// ── Helpers ─────────────────────────────────────

func (m *Manager) persist() {
	if m.store != nil {
		users := make([]User, 0, len(m.users))
		for _, u := range m.users {
			users = append(users, *u)
		}
		_ = m.store.Save(users)
	}
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func generateToken() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return "shield_" + hex.EncodeToString(b)
}
