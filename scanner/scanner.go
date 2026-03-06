package scanner

import (
	"regexp"
	"sort"
)

// PIIType identifies what kind of sensitive data was detected.
type PIIType string

const (
	PIIEmail      PIIType = "EMAIL"
	PIIPhone      PIIType = "PHONE"
	PIISSN        PIIType = "SSN"
	PIIAPIKey     PIIType = "API_KEY"
	PIICreditCard PIIType = "CREDIT_CARD"
	PIIIPAddress  PIIType = "IP_ADDR"
	PIIFilePath   PIIType = "PATH"
	PIIPersonName PIIType = "PERSON"
	PIISecret     PIIType = "SECRET"
	PIIPasswords  PIIType = "PASSWORD"
)

// Match represents a detected PII occurrence in text.
type Match struct {
	Type  PIIType
	Value string
	Start int
	End   int
}

// pattern pairs a PII type with its detection regex.
type pattern struct {
	Type PIIType
	Re   *regexp.Regexp
}

// Scanner detects PII in text using compiled regex patterns.
type Scanner struct {
	patterns []pattern
}

// New creates a Scanner with all default PII detection patterns.
func New() *Scanner {
	s := &Scanner{}

	// Order matters: more specific patterns first to avoid partial matches.
	s.add(PIISSN, `\b\d{3}-\d{2}-\d{4}\b`)
	s.add(PIICreditCard, `\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`)
	s.add(PIIEmail, `\b[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}\b`)
	s.add(PIIPhone, `\b(?:\+1[\s.-]?)?\(?\d{3}\)?[\s.-]?\d{3}[\s.-]?\d{4}\b`)

	// API keys and tokens (common patterns).
	s.add(PIIAPIKey, `\bsk-[a-zA-Z0-9]{20,}\b`)            // OpenAI
	s.add(PIIAPIKey, `\bghp_[a-zA-Z0-9]{36,}\b`)           // GitHub PAT
	s.add(PIIAPIKey, `\bgho_[a-zA-Z0-9]{36,}\b`)           // GitHub OAuth
	s.add(PIIAPIKey, `\bglpat-[a-zA-Z0-9\-_]{20,}\b`)      // GitLab PAT
	s.add(PIIAPIKey, `\bxoxb-[a-zA-Z0-9\-]+\b`)            // Slack bot token
	s.add(PIIAPIKey, `\bAIza[a-zA-Z0-9_\-]{35}\b`)         // Google API key
	s.add(PIIAPIKey, `\b[A-Z0-9]{20}:[a-zA-Z0-9/+]{40}\b`) // AWS-style key:secret
	s.add(PIISecret, `\b[A-Za-z0-9+/]{40,}={0,2}\b`)       // Base64 secrets (40+ chars)

	// Password patterns in config/env files.
	s.add(PIIPasswords, `(?i)(?:password|passwd|pwd)\s*[:=]\s*["']?([^\s"']+)["']?`)

	// File paths with usernames.
	s.add(PIIFilePath, `(?i)[A-Z]:\\Users\\[^\\]+\\[^\s"']+`)         // Windows paths
	s.add(PIIFilePath, `(?:/home/|/Users/)[a-zA-Z0-9._\-]+/[^\s"']+`) // Unix paths

	// IP addresses (private ranges are often sensitive).
	s.add(PIIIPAddress, `\b(?:10\.\d{1,3}\.\d{1,3}\.\d{1,3})\b`)               // 10.x.x.x
	s.add(PIIIPAddress, `\b(?:192\.168\.\d{1,3}\.\d{1,3})\b`)                  // 192.168.x.x
	s.add(PIIIPAddress, `\b(?:172\.(?:1[6-9]|2\d|3[01])\.\d{1,3}\.\d{1,3})\b`) // 172.16-31.x.x

	return s
}

func (s *Scanner) add(piiType PIIType, pattern string) {
	s.patterns = append(s.patterns, struct {
		Type PIIType
		Re   *regexp.Regexp
	}{
		Type: piiType,
		Re:   regexp.MustCompile(pattern),
	})
}

// Scan finds all PII matches in the given text.
// Returns matches sorted by position (earliest first).
func (s *Scanner) Scan(text string) []Match {
	var matches []Match
	seen := make(map[string]bool) // deduplicate overlapping matches

	for _, p := range s.patterns {
		locs := p.Re.FindAllStringIndex(text, -1)
		for _, loc := range locs {
			value := text[loc[0]:loc[1]]
			key := value + string(rune(loc[0])) // unique by value+position
			if seen[key] {
				continue
			}
			seen[key] = true

			matches = append(matches, Match{
				Type:  p.Type,
				Value: value,
				Start: loc[0],
				End:   loc[1],
			})
		}
	}

	// Sort by position (earliest first, longest first for ties).
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Start == matches[j].Start {
			return matches[i].End > matches[j].End
		}
		return matches[i].Start < matches[j].Start
	})

	// Remove overlapping matches (keep the first/longest).
	var filtered []Match
	lastEnd := 0
	for _, m := range matches {
		if m.Start >= lastEnd {
			filtered = append(filtered, m)
			lastEnd = m.End
		}
	}

	return filtered
}
