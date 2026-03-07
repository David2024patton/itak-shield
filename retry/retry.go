package retry

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Config holds retry and fallback settings.
type Config struct {
	MaxRetries      int      // max retry attempts per target
	BackoffMs       int      // base backoff in milliseconds (doubles each attempt)
	FallbackTargets []string // ordered list of fallback upstream URLs
}

// retryableCodes are HTTP status codes worth retrying.
var retryableCodes = map[int]bool{
	429: true, // Too Many Requests
	500: true, // Internal Server Error
	502: true, // Bad Gateway
	503: true, // Service Unavailable
	504: true, // Gateway Timeout
}

// Do executes an HTTP request with retry and fallback logic.
// It clones the request for each attempt using the saved body.
// Returns the response from whichever target succeeds, or the last error.
func Do(client *http.Client, req *http.Request, body []byte, cfg *Config) (*http.Response, error) {
	if cfg == nil {
		return client.Do(req)
	}

	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	backoffBase := time.Duration(cfg.BackoffMs) * time.Millisecond
	if backoffBase <= 0 {
		backoffBase = 500 * time.Millisecond
	}

	// Try the primary target first.
	resp, err := doWithRetries(client, req, body, maxRetries, backoffBase)
	if err == nil && !retryableCodes[resp.StatusCode] {
		return resp, nil
	}

	// Primary failed. Try fallback targets in order.
	var lastErr error
	if err != nil {
		lastErr = err
	} else {
		lastErr = fmt.Errorf("primary returned %d", resp.StatusCode)
		// Close the failed response body.
		resp.Body.Close()
	}

	for _, fallbackURL := range cfg.FallbackTargets {
		log.Printf("[iTaK Shield] Falling back to %s", fallbackURL)

		fallbackReq, cloneErr := cloneRequest(req, body, fallbackURL)
		if cloneErr != nil {
			lastErr = cloneErr
			continue
		}

		resp, err = doWithRetries(client, fallbackReq, body, maxRetries, backoffBase)
		if err == nil && !retryableCodes[resp.StatusCode] {
			log.Printf("[iTaK Shield] Fallback to %s succeeded", fallbackURL)
			return resp, nil
		}
		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("fallback %s returned %d", fallbackURL, resp.StatusCode)
			resp.Body.Close()
		}
	}

	return nil, fmt.Errorf("all targets failed: %w", lastErr)
}

// doWithRetries tries a single target with exponential backoff retries.
func doWithRetries(client *http.Client, req *http.Request, body []byte, maxRetries int, backoff time.Duration) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			wait := backoff * time.Duration(1<<(attempt-1))
			log.Printf("[iTaK Shield] Retry %d/%d after %v", attempt, maxRetries, wait)
			time.Sleep(wait)
		}

		// Clone the request body for each attempt.
		cloned, err := cloneRequest(req, body, "")
		if err != nil {
			return nil, err
		}

		resp, err := client.Do(cloned)
		if err != nil {
			lastErr = err
			continue
		}

		if !retryableCodes[resp.StatusCode] {
			return resp, nil
		}

		lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
		// Read and discard the body so the connection can be reused.
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	return nil, fmt.Errorf("max retries exhausted: %w", lastErr)
}

// cloneRequest creates a fresh request with a new body reader.
// If overrideURL is non-empty, it replaces the request URL host/scheme.
func cloneRequest(orig *http.Request, body []byte, overrideURL string) (*http.Request, error) {
	targetURL := orig.URL.String()
	if overrideURL != "" {
		// Replace scheme+host but keep path and query.
		targetURL = overrideURL + orig.URL.Path
		if orig.URL.RawQuery != "" {
			targetURL += "?" + orig.URL.RawQuery
		}
	}

	req, err := http.NewRequestWithContext(orig.Context(), orig.Method, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// Copy all headers.
	for key, vals := range orig.Header {
		for _, v := range vals {
			req.Header.Add(key, v)
		}
	}
	req.ContentLength = int64(len(body))

	return req, nil
}
