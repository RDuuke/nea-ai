package doctor

import (
	"os"
	"path/filepath"
	"testing"

	"nea-ai/internal/model"
)

func TestCheckBoolReportsOK(t *testing.T) {
	check := checkBool("id", true, "ok msg", "fail msg")
	if check.Status != CheckOK {
		t.Errorf("status = %q, want %q", check.Status, CheckOK)
	}
	if check.Message != "ok msg" {
		t.Errorf("message = %q, want %q", check.Message, "ok msg")
	}
}

func TestCheckBoolReportsFailure(t *testing.T) {
	check := checkBool("id", false, "ok msg", "fail msg")
	if check.Status != CheckFailed {
		t.Errorf("status = %q, want %q", check.Status, CheckFailed)
	}
	if check.Message != "fail msg" {
		t.Errorf("message = %q, want %q", check.Message, "fail msg")
	}
}

func TestCheckWritableSucceedsForDirectory(t *testing.T) {
	dir := t.TempDir()
	check := checkWritable("workspace.writable", dir)
	if check.Status != CheckOK {
		t.Errorf("expected ok status, got %q (%s)", check.Status, check.Message)
	}
}

func TestCheckWritableFailsForFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	check := checkWritable("workspace.writable", target)
	if check.Status != CheckFailed {
		t.Errorf("expected failed status, got %q", check.Status)
	}
}

func TestCheckWritableFailsForMissingPath(t *testing.T) {
	check := checkWritable("workspace.writable", filepath.Join(t.TempDir(), "missing"))
	if check.Status != CheckFailed {
		t.Errorf("expected failed status, got %q", check.Status)
	}
}

func TestRunForAgentReturnsReport(t *testing.T) {
	workDir := t.TempDir()
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })

	report, err := RunForAgent("test", model.AgentCodex)
	if err != nil {
		t.Fatalf("RunForAgent: %v", err)
	}
	if len(report.Checks) == 0 {
		t.Fatal("expected at least one check, got none")
	}

	// openspec.present should be failed because workDir is empty.
	var openspecCheck *Check
	for i := range report.Checks {
		if report.Checks[i].ID == "openspec.present" {
			openspecCheck = &report.Checks[i]
			break
		}
	}
	if openspecCheck == nil {
		t.Fatal("expected openspec.present check in report")
	}
	if openspecCheck.Status != CheckFailed {
		t.Errorf("openspec.present status = %q, want %q (empty workspace)", openspecCheck.Status, CheckFailed)
	}
	if report.Ready {
		t.Error("expected Ready=false on empty workspace")
	}
}
