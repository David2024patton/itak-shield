package proxy

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/David2024patton/itak-shield/audit"
	"github.com/David2024patton/itak-shield/auth"
	"github.com/David2024patton/itak-shield/cache"
	"github.com/David2024patton/itak-shield/dlp"
	"github.com/David2024patton/itak-shield/retry"
	"github.com/David2024patton/itak-shield/scanner"
	"github.com/David2024patton/itak-shield/spend"
	"github.com/David2024patton/itak-shield/tokenizer"
)

// LogEntry represents a single proxy activity event.
type LogEntry struct {
	Time    time.Time
	Type    string
	Message string
}

// LiveStatsResult holds the current proxy stats for the GUI.
type LiveStatsResult struct {
	Requests   int64            `json:"requests"`
	Redacted   int64            `json:"redacted"`
	RecentLogs []LogEntry       `json:"-"`
	CacheStats *cache.Stats     `json:"cache_stats,omitempty"`
	SpendStats *spend.Stats     `json:"spend_stats,omitempty"`
	AuthUsers  map[string]int64 `json:"auth_users,omitempty"`
	Features   FeatureFlags     `json:"features"`
}

// FeatureFlags indicates which enterprise features are active.
type FeatureFlags struct {
	Auth  bool `json:"auth"`
	Cache bool `json:"cache"`
	Retry bool `json:"retry"`
	Spend bool `json:"spend"`
	DLP   bool `json:"dlp"`
}

// Option configures the proxy server.
type Option func(*Server) error

// WithCustomRule adds an organization-specific PII detection pattern.
func WithCustomRule(name, pattern string) Option {
	return func(s *Server) error {
		return s.scanner.AddCustomRule(name, pattern)
	}
}

// WithDisabledRule disables a built-in PII detection rule.
func WithDisabledRule(name string) Option {
	return func(s *Server) error {
		s.scanner.DisableRule(name)
		return nil
	}
}

// WithAuditLogger attaches a structured audit logger to the proxy.
func WithAuditLogger(logger *audit.Logger) Option {
	return func(s *Server) error {
		s.auditLogger = logger
		return nil
	}
}

// WithAuth enables virtual API key authentication and rate limiting.
func WithAuth(manager *auth.Manager) Option {
	return func(s *Server) error {
		s.authManager = manager
		return nil
	}
}

// WithCache enables response caching.
func WithCache(c *cache.Cache) Option {
	return func(s *Server) error {
		s.cache = c
		return nil
	}
}

// WithRetry enables auto-retry with fallback routing.
func WithRetry(cfg *retry.Config) Option {
	return func(s *Server) error {
		s.retryCfg = cfg
		return nil
	}
}

// WithSpend enables token tracking and budget enforcement.
func WithSpend(tracker *spend.Tracker) Option {
	return func(s *Server) error {
		s.spendTracker = tracker
		return nil
	}
}

// WithDLP enables data loss prevention policies.
func WithDLP(policy *dlp.Policy) Option {
	return func(s *Server) error {
		s.dlpPolicy = policy
		return nil
	}
}

// Server is the iTaK Shield privacy-aware reverse proxy.
type Server struct {
	target    *url.URL
	scanner   *scanner.Scanner
	tokenizer *tokenizer.TokenMap
	verbose   bool

	// Thread-safe counters for the GUI dashboard.
	requestCount  atomic.Int64
	redactedCount atomic.Int64

	// Recent activity log (ring buffer).
	logMu      sync.Mutex
	recentLogs []LogEntry

	// Enterprise features (all nil-safe).
	auditLogger  *audit.Logger
	authManager  *auth.Manager
	cache        *cache.Cache
	retryCfg     *retry.Config
	spendTracker *spend.Tracker
	dlpPolicy    *dlp.Policy

	// Per-user request tracking for analytics.
	userStatsMu sync.Mutex
	userStats   map[string]int64
}

const maxLogEntries = 50

// New creates a new proxy server targeting the given upstream API URL.
// Accepts optional functional options for enterprise configuration.
func New(targetURL string, verbose bool, opts ...Option) (*Server, error) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid target URL %q: %w", targetURL, err)
	}

	s := &Server{
		target:     u,
		scanner:    scanner.New(),
		tokenizer:  tokenizer.New(),
		verbose:    verbose,
		recentLogs: make([]LogEntry, 0, maxLogEntries),
		userStats:  make(map[string]int64),
	}

	// Apply enterprise options.
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("proxy option error: %w", err)
		}
	}

	return s, nil
}

// LiveStats returns the current request/redaction counts and recent log entries.
func (s *Server) LiveStats() LiveStatsResult {
	s.logMu.Lock()
	logs := make([]LogEntry, len(s.recentLogs))
	copy(logs, s.recentLogs)
	s.logMu.Unlock()

	result := LiveStatsResult{
		Requests:   s.requestCount.Load(),
		Redacted:   s.redactedCount.Load(),
		RecentLogs: logs,
		Features: FeatureFlags{
			Auth:  s.authManager != nil && s.authManager.HasKeys(),
			Cache: s.cache != nil,
			Retry: s.retryCfg != nil,
			Spend: s.spendTracker != nil,
			DLP:   s.dlpPolicy != nil,
		},
	}

	// Attach enterprise stats if available.
	if s.cache != nil {
		cs := s.cache.GetStats()
		result.CacheStats = &cs
	}
	if s.spendTracker != nil {
		ss := s.spendTracker.GetStats()
		result.SpendStats = &ss
	}
	if s.authManager != nil && s.authManager.HasKeys() {
		s.userStatsMu.Lock()
		users := make(map[string]int64, len(s.userStats))
		for k, v := range s.userStats {
			users[k] = v
		}
		s.userStatsMu.Unlock()
		result.AuthUsers = users
	}

	return result
}

// addLog appends a log entry to the ring buffer.
func (s *Server) addLog(logType, message string) {
	entry := LogEntry{
		Time:    time.Now(),
		Type:    logType,
		Message: message,
	}
	s.logMu.Lock()
	defer s.logMu.Unlock()

	if len(s.recentLogs) >= maxLogEntries {
		s.recentLogs = s.recentLogs[1:]
	}
	s.recentLogs = append(s.recentLogs, entry)
}

// trackUser increments the per-user request counter.
func (s *Server) trackUser(user string) {
	s.userStatsMu.Lock()
	s.userStats[user]++
	s.userStatsMu.Unlock()
}

// ServeHTTP handles incoming requests through the enterprise middleware pipeline:
// 1. Auth (validate key, identify user, check rate limit)
// 2. Read request body
// 3. DLP (scan, check block policies)
// 4. Cache check (return cached response if hit)
// 5. PII redaction (existing)
// 6. Forward with retry/fallback
// 7. Spend tracking (parse response tokens)
// 8. Cache store (save response)
// 9. PII restoration (existing)
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.requestCount.Add(1)

	// ── Step 1: Authentication ────────────────
	user := "anonymous"
	if s.authManager != nil && s.authManager.HasKeys() {
		identity, err := s.authManager.Authenticate(r.Header.Get("Authorization"))
		if err != nil {
			s.addLog("AUTH", fmt.Sprintf("Rejected: %v", err))
			s.auditLogger.Log(audit.Event{
				EventType: "auth_fail",
				Message:   err.Error(),
				Method:    r.Method,
				Path:      r.URL.Path,
				Source:    r.RemoteAddr,
			})
			writeJSONError(w, http.StatusUnauthorized, err.Error())
			return
		}
		user = identity.User

		// Check rate limit.
		if err := s.authManager.CheckRateLimit(user); err != nil {
			s.addLog("RATE", fmt.Sprintf("Rate limited: %s", user))
			writeJSONError(w, http.StatusTooManyRequests, err.Error())
			return
		}

		// Check spend budget before processing.
		if s.spendTracker != nil {
			s.spendTracker.SetUserGroup(user, identity.Group)
			if err := s.spendTracker.CheckBudget(user); err != nil {
				s.addLog("BUDGET", fmt.Sprintf("Budget exceeded: %s", user))
				writeJSONError(w, http.StatusPaymentRequired, err.Error())
				return
			}
		}

		// Inject the real upstream API key (replace the virtual key).
		if injectKey := s.authManager.InjectKey(); injectKey != "" {
			r.Header.Set("Authorization", "Bearer "+injectKey)
		}
	}

	s.trackUser(user)

	// ── Step 2: Read request body ────────────
	var bodyBytes []byte
	if r.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}
	}

	// ── Step 3: Scan for PII ─────────────────
	s.tokenizer.Reset()
	bodyStr := string(bodyBytes)
	matches := s.scanner.Scan(bodyStr)

	// ── Step 3b: DLP policy check ────────────
	if s.dlpPolicy != nil && len(matches) > 0 {
		result := s.dlpPolicy.Evaluate(matches)
		if result.Blocked {
			s.addLog("DLP", result.Message)
			s.auditLogger.Log(audit.Event{
				EventType: "dlp_block",
				Message:   result.Message,
				Types:     result.BlockedTypes,
				Method:    r.Method,
				Path:      r.URL.Path,
				Source:    r.RemoteAddr,
			})
			writeJSONError(w, http.StatusForbidden, result.Message)
			return
		}
	}

	// ── Step 4: Cache check ──────────────────
	cacheKey := ""
	if s.cache != nil {
		cacheKey = cache.Hash(bodyBytes)
		if cached := s.cache.Get(cacheKey); cached != nil {
			s.addLog("CACHE", fmt.Sprintf("Cache hit for %s %s", r.Method, r.URL.Path))
			// Restore PII in cached response.
			restored := s.tokenizer.Restore(string(cached.Body))
			for key, values := range cached.Headers {
				for _, v := range values {
					w.Header().Add(key, v)
				}
			}
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(restored)))
			w.Header().Set("X-iTaK-Cache", "HIT")
			w.WriteHeader(cached.StatusCode)
			w.Write([]byte(restored))
			return
		}
	}

	// ── Step 5: PII redaction ────────────────
	redacted, count := s.tokenizer.Redact(bodyStr, matches)

	if count > 0 {
		s.redactedCount.Add(int64(count))
		s.addLog("REDACT", fmt.Sprintf("Redacted %d PII item(s)", count))
	}

	if s.verbose && count > 0 {
		log.Printf("[iTaK Shield] Redacted %d PII item(s) from request", count)
		stats := s.tokenizer.Stats()
		for piiType, n := range stats {
			log.Printf("  %s: %d", piiType, n)
		}
	}

	// Audit logging.
	if count > 0 {
		var piiTypes []string
		seen := make(map[string]bool)
		for _, m := range matches {
			t := string(m.Type)
			if !seen[t] {
				seen[t] = true
				piiTypes = append(piiTypes, t)
			}
		}
		s.auditLogger.Log(audit.Event{
			EventType: "redact",
			Items:     count,
			Types:     piiTypes,
			Method:    r.Method,
			Path:      r.URL.Path,
			Source:    r.RemoteAddr,
		})
	} else {
		s.auditLogger.Log(audit.Event{
			EventType: "pass",
			Method:    r.Method,
			Path:      r.URL.Path,
			Source:    r.RemoteAddr,
		})
	}

	// ── Step 6: Build and forward request ────
	upstreamURL := *s.target
	upstreamURL.Path = r.URL.Path
	upstreamURL.RawQuery = r.URL.RawQuery

	redactedBytes := []byte(redacted)
	upReq, err := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL.String(), bytes.NewReader(redactedBytes))
	if err != nil {
		http.Error(w, "failed to create upstream request", http.StatusInternalServerError)
		return
	}

	// Copy headers (except Host).
	for key, values := range r.Header {
		if strings.EqualFold(key, "Host") {
			continue
		}
		for _, v := range values {
			upReq.Header.Add(key, v)
		}
	}
	upReq.Header.Set("Host", s.target.Host)
	upReq.ContentLength = int64(len(redactedBytes))

	// Forward with retry/fallback or direct.
	client := &http.Client{}
	var resp *http.Response

	if s.retryCfg != nil {
		resp, err = retry.Do(client, upReq, redactedBytes, s.retryCfg)
	} else {
		resp, err = client.Do(upReq)
	}

	if err != nil {
		s.addLog("ERROR", fmt.Sprintf("Upstream failed: %v", err))
		s.auditLogger.Log(audit.Event{
			EventType: "error",
			Message:   fmt.Sprintf("Upstream failed: %v", err),
			Method:    r.Method,
			Path:      r.URL.Path,
		})
		http.Error(w, fmt.Sprintf("upstream request failed: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Read the response body, handling gzip if needed.
	var respBody []byte
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, gzErr := gzip.NewReader(resp.Body)
		if gzErr != nil {
			http.Error(w, "failed to decompress response", http.StatusBadGateway)
			return
		}
		respBody, err = io.ReadAll(gz)
		gz.Close()
	} else {
		respBody, err = io.ReadAll(resp.Body)
	}
	if err != nil {
		http.Error(w, "failed to read upstream response", http.StatusBadGateway)
		return
	}

	// ── Step 7: Spend tracking ───────────────
	if s.spendTracker != nil {
		s.spendTracker.TrackResponse(user, respBody)
	}

	// ── Step 8: Cache store ──────────────────
	if s.cache != nil && cacheKey != "" && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		headers := make(map[string][]string)
		for key, values := range resp.Header {
			lower := strings.ToLower(key)
			if lower == "content-length" || lower == "content-encoding" || lower == "transfer-encoding" {
				continue
			}
			headers[key] = values
		}
		s.cache.Set(cacheKey, &cache.Entry{
			Body:       respBody,
			Headers:    headers,
			StatusCode: resp.StatusCode,
			CreatedAt:  time.Now(),
		})
	}

	// ── Step 9: Restore PII in response ──────
	restored := s.tokenizer.Restore(string(respBody))

	if s.verbose && count > 0 {
		log.Printf("[iTaK Shield] Restored tokens in response (%d bytes)", len(restored))
	}

	if count > 0 {
		s.addLog("RESTORE", fmt.Sprintf("Restored %d token(s) in response (%d bytes)", count, len(restored)))
	} else {
		s.addLog("PASS", fmt.Sprintf("Proxied %s %s (%d bytes)", r.Method, r.URL.Path, len(restored)))
	}

	// Copy response headers.
	for key, values := range resp.Header {
		lower := strings.ToLower(key)
		if lower == "content-length" || lower == "content-encoding" || lower == "transfer-encoding" {
			continue
		}
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}

	if s.cache != nil {
		w.Header().Set("X-iTaK-Cache", "MISS")
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(restored)))
	w.WriteHeader(resp.StatusCode)
	w.Write([]byte(restored))
}

// writeJSONError writes a JSON error response matching OpenAI error format.
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"message": msg,
			"type":    "itak_shield_error",
			"code":    status,
		},
	})
}
