package scanner

import (
	"testing"
)

func TestScanEmail(t *testing.T) {
	s := New()
	matches := s.Scan("Contact john.doe@example.com for help")
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].Type != PIIEmail {
		t.Errorf("type: got %s, want EMAIL", matches[0].Type)
	}
	if matches[0].Value != "john.doe@example.com" {
		t.Errorf("value: got %q", matches[0].Value)
	}
	t.Logf("Email detection: PASSED")
}

func TestScanSSN(t *testing.T) {
	s := New()
	matches := s.Scan("SSN is 123-45-6789 on file")
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].Type != PIISSN {
		t.Errorf("type: got %s, want SSN", matches[0].Type)
	}
	if matches[0].Value != "123-45-6789" {
		t.Errorf("value: got %q", matches[0].Value)
	}
	t.Logf("SSN detection: PASSED")
}

func TestScanAPIKey(t *testing.T) {
	s := New()
	matches := s.Scan("key is sk-abcdefghij1234567890XY end")
	found := false
	for _, m := range matches {
		if m.Type == PIIAPIKey {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("failed to detect OpenAI API key")
	}
	t.Logf("API key detection: PASSED")
}

func TestScanFilePath(t *testing.T) {
	s := New()
	matches := s.Scan(`Check C:\Users\David\Projects\secret\app.py for bugs`)
	found := false
	for _, m := range matches {
		if m.Type == PIIFilePath {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("failed to detect Windows file path")
	}
	t.Logf("File path detection: PASSED")
}

func TestScanIPAddress(t *testing.T) {
	s := New()
	matches := s.Scan("Server at 192.168.1.100 is down")
	found := false
	for _, m := range matches {
		if m.Type == PIIIPAddress {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("failed to detect private IP address")
	}
	t.Logf("IP address detection: PASSED")
}

func TestScanPhone(t *testing.T) {
	s := New()
	matches := s.Scan("Call me at (555) 123-4567")
	found := false
	for _, m := range matches {
		if m.Type == PIIPhone {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("failed to detect phone number")
	}
	t.Logf("Phone detection: PASSED")
}

func TestScanMultiplePII(t *testing.T) {
	s := New()
	text := "Email: user@test.com, SSN: 999-88-7777, IP: 10.0.0.1"
	matches := s.Scan(text)

	types := make(map[PIIType]bool)
	for _, m := range matches {
		types[m.Type] = true
	}

	if !types[PIIEmail] {
		t.Error("missing EMAIL")
	}
	if !types[PIISSN] {
		t.Error("missing SSN")
	}
	if !types[PIIIPAddress] {
		t.Error("missing IP_ADDR")
	}
	t.Logf("Multiple PII: PASSED (%d matches, %d types)", len(matches), len(types))
}

func TestScanNoFalsePositives(t *testing.T) {
	s := New()
	// Normal text should not trigger detections.
	matches := s.Scan("The quick brown fox jumps over the lazy dog. Version 3.14 is released.")
	if len(matches) != 0 {
		t.Errorf("false positives: got %d matches in clean text", len(matches))
		for _, m := range matches {
			t.Logf("  %s: %q", m.Type, m.Value)
		}
	}
	t.Logf("No false positives: PASSED")
}
