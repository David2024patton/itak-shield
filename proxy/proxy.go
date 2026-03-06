package proxy

import (
	"bytes"
	"compress/gzip"
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
	"github.com/David2024patton/itak-shield/scanner"
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
	Requests   int64
	Redacted   int64
	RecentLogs []LogEntry
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

	// Enterprise: structured audit logger (nil-safe).
	auditLogger *audit.Logger
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

	return LiveStatsResult{
		Requests:   s.requestCount.Load(),
		Redacted:   s.redactedCount.Load(),
		RecentLogs: logs,
	}
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
		// Shift older entries out.
		s.recentLogs = s.recentLogs[1:]
	}
	s.recentLogs = append(s.recentLogs, entry)
}

// ServeHTTP handles incoming requests by:
// 1. Reading the request body
// 2. Scanning for PII and replacing with tokens
// 3. Forwarding the sanitized request to the upstream API
// 4. Reading the response and restoring tokens to real values
// 5. Returning the de-tokenized response to the caller
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Increment request counter.
	s.requestCount.Add(1)

	// Read the request body.
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

	// Reset tokenizer for each request to prevent cross-contamination.
	s.tokenizer.Reset()

	// Scan and redact PII from the request body.
	bodyStr := string(bodyBytes)
	matches := s.scanner.Scan(bodyStr)
	redacted, count := s.tokenizer.Redact(bodyStr, matches)

	// Track redacted count.
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

	// Enterprise: structured audit logging (metadata only, never PII).
	if count > 0 {
		var piiTypes []string
		for _, m := range matches {
			piiTypes = append(piiTypes, string(m.Type))
		}
		// Deduplicate types.
		seen := make(map[string]bool)
		var unique []string
		for _, t := range piiTypes {
			if !seen[t] {
				seen[t] = true
				unique = append(unique, t)
			}
		}
		s.auditLogger.Log(audit.Event{
			EventType: "redact",
			Items:     count,
			Types:     unique,
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

	// Build the upstream request.
	upstreamURL := *s.target
	upstreamURL.Path = r.URL.Path
	upstreamURL.RawQuery = r.URL.RawQuery

	upReq, err := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL.String(), bytes.NewReader([]byte(redacted)))
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
	// Fix content-length for the redacted body.
	upReq.ContentLength = int64(len(redacted))

	// Forward to upstream.
	client := &http.Client{}
	resp, err := client.Do(upReq)
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

	// Restore PII tokens in the response.
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
		// Skip content-length and encoding since we may have changed the body.
		lower := strings.ToLower(key)
		if lower == "content-length" || lower == "content-encoding" || lower == "transfer-encoding" {
			continue
		}
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(restored)))
	w.WriteHeader(resp.StatusCode)
	w.Write([]byte(restored))
}
