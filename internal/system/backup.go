package system

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BackupFile writes a timestamped copy of path to "<path>.nea-ai.<ts>.bak".
// Returns ("", nil) when path does not exist.
func BackupFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	backupPath := fmt.Sprintf("%s.nea-ai.%s.bak", path, time.Now().UTC().Format("20060102150405"))
	if err := os.MkdirAll(filepath.Dir(backupPath), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(backupPath, data, 0o644); err != nil {
		return "", err
	}
	return backupPath, nil
}
