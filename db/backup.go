package db

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// BackupConfig holds settings for automatic database backups.
type BackupConfig struct {
	Enabled     bool   `json:"enabled"`
	IntervalH   int    `json:"interval_hours"` // default 12
	MaxBackups  int    `json:"max_backups"`    // default 5
	BackupDir   string `json:"backup_dir"`     // defaults to same dir as DB
}

// DefaultBackupConfig returns the default backup configuration.
func DefaultBackupConfig() BackupConfig {
	return BackupConfig{
		Enabled:    true,
		IntervalH:  12,
		MaxBackups: 5,
	}
}

// Backup creates a copy of the SQLite database file.
// Returns the path to the backup file.
func (d *DB) Backup(backupDir string) (string, error) {
	if d == nil || d.conn == nil {
		return "", fmt.Errorf("database not open")
	}

	if backupDir == "" {
		backupDir = filepath.Dir(d.path)
	}
	if err := os.MkdirAll(backupDir, 0700); err != nil {
		return "", fmt.Errorf("cannot create backup dir: %w", err)
	}

	// Force a WAL checkpoint so all data is in the main DB file.
	d.mu.Lock()
	_, err := d.conn.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	d.mu.Unlock()
	if err != nil {
		return "", fmt.Errorf("wal checkpoint: %w", err)
	}

	ts := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("shield-backup-%s.db", ts)
	backupPath := filepath.Join(backupDir, backupName)

	// Copy file.
	src, err := os.Open(d.path)
	if err != nil {
		return "", fmt.Errorf("open source: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("create backup: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(backupPath)
		return "", fmt.Errorf("copy: %w", err)
	}

	return backupPath, nil
}

// PruneBackups removes old backups, keeping only the latest maxKeep.
func (d *DB) PruneBackups(backupDir string, maxKeep int) error {
	if backupDir == "" {
		backupDir = filepath.Dir(d.path)
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return err
	}

	var backups []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "shield-backup-") && strings.HasSuffix(e.Name(), ".db") {
			backups = append(backups, filepath.Join(backupDir, e.Name()))
		}
	}

	// Sort oldest first.
	sort.Strings(backups)

	if len(backups) <= maxKeep {
		return nil
	}

	// Remove the oldest.
	toRemove := backups[:len(backups)-maxKeep]
	for _, path := range toRemove {
		os.Remove(path)
	}

	return nil
}

// ListBackups returns all backup files sorted newest first.
func (d *DB) ListBackups(backupDir string) ([]BackupInfo, error) {
	if backupDir == "" {
		backupDir = filepath.Dir(d.path)
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var backups []BackupInfo
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "shield-backup-") && strings.HasSuffix(e.Name(), ".db") {
			info, err := e.Info()
			if err != nil {
				continue
			}
			backups = append(backups, BackupInfo{
				Name:      e.Name(),
				Path:      filepath.Join(backupDir, e.Name()),
				Size:      info.Size(),
				CreatedAt: info.ModTime(),
			})
		}
	}

	// Sort newest first.
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

// RestoreBackup replaces the current database with a backup.
// The caller should stop all activity before calling this.
func (d *DB) RestoreBackup(backupPath string) error {
	if d == nil || d.conn == nil {
		return fmt.Errorf("database not open")
	}

	// Verify backup file exists.
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup not found: %w", err)
	}

	// Close current connection.
	d.mu.Lock()
	defer d.mu.Unlock()

	d.conn.Close()

	// Copy backup over current DB.
	src, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("open backup: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(d.path)
	if err != nil {
		return fmt.Errorf("overwrite db: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	// Reopen connection.
	conn, err := openConn(d.path)
	if err != nil {
		return fmt.Errorf("reopen: %w", err)
	}
	d.conn = conn

	return nil
}

// BackupInfo holds metadata about a backup file.
type BackupInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

// openConn opens a connection with standard pragmas.
func openConn(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	_, _ = conn.Exec("PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;")
	return conn, nil
}

// StartAutoBackup runs periodic backups in a goroutine.
// Returns a stop channel - send to it to stop the backup loop.
func (d *DB) StartAutoBackup(cfg BackupConfig) chan struct{} {
	stop := make(chan struct{})

	if !cfg.Enabled || cfg.IntervalH <= 0 {
		return stop
	}

	interval := time.Duration(cfg.IntervalH) * time.Hour
	maxKeep := cfg.MaxBackups
	if maxKeep <= 0 {
		maxKeep = 5
	}
	backupDir := cfg.BackupDir

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				path, err := d.Backup(backupDir)
				if err != nil {
					fmt.Printf("[shield] auto-backup failed: %v\n", err)
					continue
				}
				fmt.Printf("[shield] auto-backup created: %s\n", path)
				_ = d.PruneBackups(backupDir, maxKeep)
			}
		}
	}()

	return stop
}
