package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Event represents a single audit log entry.
// Only metadata is logged - never actual PII values.
type Event struct {
	Timestamp string   `json:"ts"`
	EventType string   `json:"event"`
	Items     int      `json:"items,omitempty"`
	Types     []string `json:"types,omitempty"`
	Method    string   `json:"method,omitempty"`
	Path      string   `json:"path,omitempty"`
	Source    string   `json:"src,omitempty"`
	Message   string   `json:"msg,omitempty"`
}

// Logger writes structured JSON Lines audit events to a file.
// Safe for concurrent use. Supports automatic file rotation by size.
type Logger struct {
	mu       sync.Mutex
	file     *os.File
	path     string
	maxSize  int64 // bytes
	maxFiles int
	written  int64
}

// New creates a new audit Logger. If path is empty or maxSize is 0,
// returns a nil-safe logger that silently discards all events.
func New(path string, maxSizeMB int, maxFiles int) (*Logger, error) {
	if path == "" {
		return nil, nil
	}

	// Ensure the directory exists.
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("audit: cannot create log directory %q: %w", dir, err)
		}
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("audit: cannot open log file %q: %w", path, err)
	}

	// Get current file size for rotation tracking.
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	maxBytes := int64(maxSizeMB) * 1024 * 1024
	if maxBytes == 0 {
		maxBytes = 100 * 1024 * 1024 // 100MB default
	}

	return &Logger{
		file:     f,
		path:     path,
		maxSize:  maxBytes,
		maxFiles: maxFiles,
		written:  info.Size(),
	}, nil
}

// Log writes an audit event. Safe to call on a nil Logger (no-op).
func (l *Logger) Log(event Event) {
	if l == nil {
		return
	}

	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	data = append(data, '\n')

	l.mu.Lock()
	defer l.mu.Unlock()

	// Rotate if needed.
	if l.written+int64(len(data)) > l.maxSize {
		l.rotate()
	}

	n, _ := l.file.Write(data)
	l.written += int64(n)
}

// rotate renames the current log and opens a new one.
// Must be called with l.mu held.
func (l *Logger) rotate() {
	l.file.Close()

	// Shift existing rotated files: .9 -> .10 (deleted), .8 -> .9, etc.
	for i := l.maxFiles - 1; i >= 1; i-- {
		old := fmt.Sprintf("%s.%d", l.path, i)
		new := fmt.Sprintf("%s.%d", l.path, i+1)
		os.Rename(old, new)
	}
	// Delete the oldest if it exceeds maxFiles.
	os.Remove(fmt.Sprintf("%s.%d", l.path, l.maxFiles))

	// Current file becomes .1.
	os.Rename(l.path, l.path+".1")

	// Open a fresh file.
	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	l.file = f
	l.written = 0
}

// Close flushes and closes the audit log. Safe to call on nil.
func (l *Logger) Close() error {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}
