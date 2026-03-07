package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Store is the interface for persisting user/token data.
type Store interface {
	Load() ([]User, error)
	Save(users []User) error
}

// FileStore persists users to a JSON file on disk.
type FileStore struct {
	mu   sync.Mutex
	path string
}

// NewFileStore creates a FileStore that reads/writes to the given path.
// The parent directory is created automatically if it doesn't exist.
func NewFileStore(path string) *FileStore {
	dir := filepath.Dir(path)
	_ = os.MkdirAll(dir, 0700)
	return &FileStore{path: path}
}

// Load reads users from the JSON file. Returns empty slice if file doesn't exist.
func (s *FileStore) Load() ([]User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []User{}, nil
		}
		return nil, err
	}

	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}
	return users, nil
}

// Save writes the full user list to the JSON file atomically.
func (s *FileStore) Save(users []User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	// Write to temp file then rename for atomic operation.
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
