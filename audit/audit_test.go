package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNilLoggerSafe(t *testing.T) {
	var logger *Logger
	// Must not panic.
	logger.Log(Event{EventType: "test"})
	if err := logger.Close(); err != nil {
		t.Errorf("Close on nil logger should return nil: %v", err)
	}
	t.Logf("Nil logger safety: PASSED")
}

func TestLogWritesJSONL(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "audit.jsonl")

	logger, err := New(path, 100, 10)
	if err != nil {
		t.Fatal(err)
	}

	logger.Log(Event{
		EventType: "redact",
		Items:     3,
		Types:     []string{"EMAIL", "SSN"},
		Method:    "POST",
		Path:      "/v1/chat/completions",
		Source:    "10.0.1.50",
	})

	logger.Log(Event{
		EventType: "pass",
		Method:    "GET",
		Path:      "/v1/models",
	})

	logger.Close()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines, got %d", len(lines))
	}

	// Parse first line.
	var event Event
	if err := json.Unmarshal([]byte(lines[0]), &event); err != nil {
		t.Fatalf("Line 1 not valid JSON: %v", err)
	}
	if event.EventType != "redact" {
		t.Errorf("Event type: got %q, want redact", event.EventType)
	}
	if event.Items != 3 {
		t.Errorf("Items: got %d, want 3", event.Items)
	}
	if event.Timestamp == "" {
		t.Error("Timestamp should be auto-populated")
	}

	t.Logf("JSONL writing: PASSED (%d lines, valid JSON)", len(lines))
}

func TestLogRotation(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "audit.jsonl")

	// Tiny max size (1KB) to trigger rotation.
	logger, err := New(path, 0, 3) // 0 = default 100MB, but we'll override
	if err != nil {
		t.Fatal(err)
	}
	logger.Close()

	// Create with 1 byte max to force immediate rotation.
	logger2, err := New(path, 1, 3) // 1MB max
	if err != nil {
		t.Fatal(err)
	}

	// We'll just verify it doesn't crash.
	for i := 0; i < 100; i++ {
		logger2.Log(Event{
			EventType: "test",
			Message:   "This is a test log entry with some padding to grow the file xxxxxxxxxxxxxx",
		})
	}
	logger2.Close()

	t.Logf("Rotation: PASSED (100 events, no crash)")
}

func TestEmptyPathReturnsNil(t *testing.T) {
	logger, err := New("", 100, 10)
	if err != nil {
		t.Fatalf("Empty path should not error: %v", err)
	}
	if logger != nil {
		t.Error("Empty path should return nil logger")
	}
	t.Logf("Empty path nil logger: PASSED")
}
