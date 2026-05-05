package brain

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"nea-ai/internal/model"
)

func TestInstallRejectsUnsupportedAgent(t *testing.T) {
	_, err := Install(InstallOptions{Agent: model.AgentCursor, Component: model.ComponentBrain})
	if err == nil {
		t.Fatal("expected error for cursor agent, got nil")
	}
	if !strings.Contains(err.Error(), "cursor") {
		t.Errorf("error %q does not mention cursor", err.Error())
	}
}

func TestUninstallRejectsUnsupportedAgent(t *testing.T) {
	_, err := Uninstall(UninstallOptions{Agent: model.AgentVSCode, Component: model.ComponentBrain})
	if err == nil {
		t.Fatal("expected error for vscode agent, got nil")
	}
}

func TestInstallRejectsUnsupportedComponent(t *testing.T) {
	_, err := Install(InstallOptions{Agent: model.AgentCodex, Component: model.ComponentFlow})
	if err == nil {
		t.Fatal("expected error for non-brain component, got nil")
	}
}

func TestSiblingCandidatesProducesExpectedPaths(t *testing.T) {
	work := filepath.Join(string(filepath.Separator)+"tmp", "demo", "nea-ai")
	got := siblingCandidates(work)
	if len(got) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(got))
	}
	binName := "neabrain"
	if runtime.GOOS == "windows" {
		binName = "neabrain.exe"
	}
	wantSuffixes := []string{
		filepath.Join("nea-brain", binName),
		filepath.Join("neabrain", binName),
	}
	for i, candidate := range got {
		if !strings.HasSuffix(candidate, wantSuffixes[i]) {
			t.Errorf("candidate[%d]=%q does not end with %q", i, candidate, wantSuffixes[i])
		}
	}
}

func TestResolveNeaBrainFindsSiblingBinary(t *testing.T) {
	parent := t.TempDir()
	workDir := filepath.Join(parent, "nea-ai")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir workDir: %v", err)
	}
	siblingDir := filepath.Join(parent, "nea-brain")
	if err := os.MkdirAll(siblingDir, 0o755); err != nil {
		t.Fatalf("mkdir sibling: %v", err)
	}
	binName := "neabrain"
	if runtime.GOOS == "windows" {
		binName = "neabrain.exe"
	}
	binPath := filepath.Join(siblingDir, binName)
	if err := os.WriteFile(binPath, []byte("stub"), 0o755); err != nil {
		t.Fatalf("write stub bin: %v", err)
	}

	t.Setenv("PATH", "")

	resolved, err := ResolveNeaBrain(workDir)
	if err != nil {
		t.Fatalf("ResolveNeaBrain: %v", err)
	}
	if resolved != binPath {
		t.Errorf("resolved=%q, want %q", resolved, binPath)
	}
}

func TestResolveNeaBrainErrorsWhenMissing(t *testing.T) {
	parent := t.TempDir()
	workDir := filepath.Join(parent, "nea-ai")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir workDir: %v", err)
	}
	t.Setenv("PATH", "")

	_, err := ResolveNeaBrain(workDir)
	if err == nil {
		t.Fatal("expected error when neabrain missing")
	}
	if !strings.Contains(err.Error(), "neabrain") {
		t.Errorf("error %q should mention neabrain", err.Error())
	}
}
