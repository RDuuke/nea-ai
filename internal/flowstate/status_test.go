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

func TestQuickCreatesBlueprintAndUpdatesStatus(t *testing.T) {
	dir := initializedOpenSpec(t)

	result, err := Quick(dir, QuickOptions{
		Name:      "fix-readme",
		Title:     "adjust readme",
		Objective: "Improve public documentation.",
		Files:     []string{"README.md"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "ok" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if _, err := os.Stat(filepath.Join(dir, "openspec", "changes", "fix-readme", "quick.md")); err != nil {
		t.Fatal(err)
	}

	status, err := Build(dir)
	if err != nil {
		t.Fatal(err)
	}
	if status.Change != "fix-readme" || status.CurrentPhase != "QUICK" || !status.AwaitingApproval {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func TestQuickRejectsActiveChange(t *testing.T) {
	dir := initializedOpenSpec(t)
	if _, err := Quick(dir, QuickOptions{Name: "first-change"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Quick(dir, QuickOptions{Name: "second-change"}); err == nil {
		t.Fatal("expected active change error")
	}
}

func TestExploreCreatesArtifactAndActivatesChange(t *testing.T) {
	dir := initializedOpenSpec(t)

	result, err := Explore(dir, PhaseOptions{
		Name:      "add-flow-cli",
		Objective: "Map flow phases to CLI commands.",
		Files:     []string{"internal/app/app.go"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.CurrentPhase != "EXPLORE" || len(result.Artifacts) != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if _, err := os.Stat(filepath.Join(dir, "openspec", "changes", "add-flow-cli", "exploration.md")); err != nil {
		t.Fatal(err)
	}

	status, err := Build(dir)
	if err != nil {
		t.Fatal(err)
	}
	if status.Change != "add-flow-cli" || status.CurrentPhase != "EXPLORE" || status.AwaitingApproval {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func TestProposeCreatesArtifactAndRequiresApproval(t *testing.T) {
	dir := initializedOpenSpec(t)
	if _, err := Explore(dir, PhaseOptions{Name: "add-flow-cli"}); err != nil {
		t.Fatal(err)
	}

	result, err := Propose(dir, PhaseOptions{
		Name:    "add-flow-cli",
		Summary: "Add bounded native CLI phase commands.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.CurrentPhase != "PROPOSE" || len(result.Artifacts) != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}

	status, err := Build(dir)
	if err != nil {
		t.Fatal(err)
	}
	if status.Change != "add-flow-cli" || status.CurrentPhase != "PROPOSE" || !status.AwaitingApproval {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func TestContinueReportsNextCommand(t *testing.T) {
	dir := initializedOpenSpec(t)
	if _, err := Explore(dir, PhaseOptions{Name: "add-flow-cli"}); err != nil {
		t.Fatal(err)
	}

	result, err := Continue(dir)
	if err != nil {
		t.Fatal(err)
	}
	if result.NextPhase != "PROPOSE" {
		t.Fatalf("next phase = %q, want PROPOSE", result.NextPhase)
	}
	if result.NextRecommended == "" {
		t.Fatal("expected next recommendation")
	}
}

func TestVerifyWritesReportAndCompletesChange(t *testing.T) {
	dir := initializedOpenSpec(t)
	if _, err := Explore(dir, PhaseOptions{Name: "add-flow-cli"}); err != nil {
		t.Fatal(err)
	}

	result, err := Verify(dir, PhaseOptions{
		Name:     "add-flow-cli",
		Summary:  "All checks passed.",
		Commands: []string{"go test ./..."},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.CurrentPhase != "VERIFY" || len(result.Artifacts) != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if _, err := os.Stat(filepath.Join(dir, "openspec", "changes", "add-flow-cli", "verify-report.md")); err != nil {
		t.Fatal(err)
	}

	status, err := Build(dir)
	if err != nil {
		t.Fatal(err)
	}
	if status.Change != "add-flow-cli" || status.CurrentPhase != "VERIFY" || !status.Completed {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func initializedOpenSpec(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	statusDir := filepath.Join(dir, "openspec", "changes")
	if err := os.MkdirAll(statusDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "openspec", "config.yaml"), []byte("project: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	status := `change: ""
current_phase: INIT
pending_tasks: []
awaiting_approval: false
completed: false
modified_artifacts: []
notes: "initialized"
`
	if err := os.WriteFile(filepath.Join(statusDir, ".status.yaml"), []byte(status), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}
