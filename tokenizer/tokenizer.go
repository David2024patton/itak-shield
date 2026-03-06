package tokenizer

import (
	"fmt"
	"strings"
	"sync"

	"github.com/David2024patton/itak-shield/scanner"
)

// TokenMap manages bidirectional mapping between real PII values and placeholder tokens.
// It is safe for concurrent use and lives only in memory (never persisted).
type TokenMap struct {
	mu       sync.RWMutex
	counters map[scanner.PIIType]int // per-type counter for unique IDs
	toToken  map[string]string       // real value -> token
	toReal   map[string]string       // token -> real value
}

// New creates an empty TokenMap.
func New() *TokenMap {
	return &TokenMap{
		counters: make(map[scanner.PIIType]int),
		toToken:  make(map[string]string),
		toReal:   make(map[string]string),
	}
}

// GetOrCreate returns the placeholder token for a real value.
// If the value has been seen before, it returns the same token.
// Otherwise, it creates a new one like [EMAIL_1], [SSN_2], etc.
func (tm *TokenMap) GetOrCreate(piiType scanner.PIIType, realValue string) string {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if token, ok := tm.toToken[realValue]; ok {
		return token
	}

	tm.counters[piiType]++
	token := fmt.Sprintf("[%s_%d]", piiType, tm.counters[piiType])

	tm.toToken[realValue] = token
	tm.toReal[token] = realValue

	return token
}

// Redact replaces all PII in text with placeholder tokens.
// Returns the redacted text and the number of replacements made.
func (tm *TokenMap) Redact(text string, matches []scanner.Match) (string, int) {
	if len(matches) == 0 {
		return text, 0
	}

	// Build result by walking through the text and replacing matches.
	var result strings.Builder
	result.Grow(len(text))

	lastPos := 0
	count := 0

	for _, m := range matches {
		if m.Start < lastPos {
			continue // skip overlapping
		}

		// Copy text before this match.
		result.WriteString(text[lastPos:m.Start])

		// Replace with token.
		token := tm.GetOrCreate(m.Type, m.Value)
		result.WriteString(token)

		lastPos = m.End
		count++
	}

	// Copy remaining text.
	result.WriteString(text[lastPos:])

	return result.String(), count
}

// Restore replaces all placeholder tokens with their original real values.
func (tm *TokenMap) Restore(text string) string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	result := text
	for token, real := range tm.toReal {
		result = strings.ReplaceAll(result, token, real)
	}
	return result
}

// Stats returns the number of unique PII values tracked.
func (tm *TokenMap) Stats() map[scanner.PIIType]int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	out := make(map[scanner.PIIType]int, len(tm.counters))
	for k, v := range tm.counters {
		out[k] = v
	}
	return out
}

// Reset clears all mappings. Call this between unrelated requests
// to prevent cross-contamination.
func (tm *TokenMap) Reset() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.counters = make(map[scanner.PIIType]int)
	tm.toToken = make(map[string]string)
	tm.toReal = make(map[string]string)
}
