package flowstate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildReadsOpenSpecStatus(t *testing.T) {
	dir := t.TempDir()
	statusDir := filepath.Join(dir, "openspec", "changes")
	if err := os.MkdirAll(statusDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "openspec", "config.yaml"), []byte("project: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	content := `change: "add-thing"
current_phase: APPLY
pending_tasks: ["task-1", "task-2"]
awaiting_approval: false
completed: false
modified_artifacts: []
notes: "working"
`
	if err := os.WriteFile(filepath.Join(statusDir, ".status.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	status, err := Build(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Present || !status.ConfigPresent || !status.StatusPresent {
		t.Fatal("expected openspec state to be present")
	}
	if status.Change != "add-thing" || status.CurrentPhase != "APPLY" {
		t.Fatalf("unexpected status: %+v", status)
	}
	if len(status.PendingTasks) != 2 {
		t.Fatalf("unexpected pending tasks: %+v", status.PendingTasks)
	}
	if status.NextRecommended != "Continue APPLY for pending tasks." {
		t.Fatalf("unexpected next recommendation: %s", status.NextRecommended)
	}
}

func TestBuildMissingStatusRecommendsInit(t *testing.T) {
	status, err := Build(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if status.StatusPresent {
		t.Fatal("status should not be present")
	}
	if status.NextRecommended == "" {
		t.Fatal("expected recommendation")
	}
}
