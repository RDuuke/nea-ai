package system

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackupFileMissingPathReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "nope.txt")

	backupPath, err := BackupFile(target)
	if err != nil {
		t.Fatalf("BackupFile returned error: %v", err)
	}
	if backupPath != "" {
		t.Fatalf("expected empty backup path for missing file, got %q", backupPath)
	}
}

func TestBackupFileCopiesContents(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "config.toml")
	want := []byte("hello = \"world\"\n")
	if err := os.WriteFile(target, want, 0o644); err != nil {
		t.Fatalf("seed target: %v", err)
	}

	backupPath, err := BackupFile(target)
	if err != nil {
		t.Fatalf("BackupFile: %v", err)
	}
	if backupPath == "" {
		t.Fatal("expected non-empty backup path")
	}
	if !strings.HasPrefix(backupPath, target+".nea-ai.") {
		t.Errorf("backup path %q does not have nea-ai prefix on target", backupPath)
	}
	if !strings.HasSuffix(backupPath, ".bak") {
		t.Errorf("backup path %q missing .bak suffix", backupPath)
	}

	got, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("backup contents = %q, want %q", got, want)
	}

	original, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read original: %v", err)
	}
	if string(original) != string(want) {
		t.Errorf("original mutated to %q, want %q", original, want)
	}
}

func TestBackupFileCreatesParentDirIfMissing(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(target, []byte("{}"), 0o644); err != nil {
		t.Fatalf("seed target: %v", err)
	}

	backupPath, err := BackupFile(target)
	if err != nil {
		t.Fatalf("BackupFile: %v", err)
	}
	if _, err := os.Stat(filepath.Dir(backupPath)); err != nil {
		t.Fatalf("parent dir missing: %v", err)
	}
}
