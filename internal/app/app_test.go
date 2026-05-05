package app

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestRunNoArgsPrintsHelp(t *testing.T) {
	var stdout bytes.Buffer
	if err := Run(nil, &stdout); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Errorf("expected help to contain Usage section, got %q", stdout.String())
	}
}

func TestRunVersionPrintsVersion(t *testing.T) {
	previous := Version
	Version = "v0.0.0-test"
	t.Cleanup(func() { Version = previous })

	var stdout bytes.Buffer
	if err := Run([]string{"version"}, &stdout); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "nea-ai v0.0.0-test" {
		t.Errorf("version output = %q, want %q", got, "nea-ai v0.0.0-test")
	}
}

func TestRunUnknownCommandFails(t *testing.T) {
	err := Run([]string{"definitely-not-a-command"}, io.Discard)
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("error %q should mention unknown command", err.Error())
	}
}

func TestRunInstallRejectsKnownButNotInstallable(t *testing.T) {
	err := Run([]string{"install", "--agent", "cursor", "--components", "brain"}, io.Discard)
	if err == nil {
		t.Fatal("expected error for cursor agent")
	}
	if !strings.Contains(err.Error(), "not installable") {
		t.Errorf("error %q should mention not installable", err.Error())
	}
}

func TestRunInstallRejectsUnknownAgent(t *testing.T) {
	err := Run([]string{"install", "--agent", "bogus", "--components", "brain"}, io.Discard)
	if err == nil {
		t.Fatal("expected error for bogus agent")
	}
}

func TestRunUninstallRejectsKnownButNotInstallable(t *testing.T) {
	err := Run([]string{"uninstall", "--agent", "vscode", "--components", "brain"}, io.Discard)
	if err == nil {
		t.Fatal("expected error for vscode agent")
	}
}

func TestRunStatusRejectsBogusAgent(t *testing.T) {
	err := Run([]string{"status", "--agent", "totally-unknown"}, io.Discard)
	if err == nil {
		t.Fatal("expected error for unknown agent")
	}
	if !strings.Contains(err.Error(), "unknown agent") {
		t.Errorf("error %q should mention unknown agent", err.Error())
	}
}

func TestRunDoctorRejectsBogusAgent(t *testing.T) {
	err := Run([]string{"doctor", "--agent", "totally-unknown"}, io.Discard)
	if err == nil {
		t.Fatal("expected error for unknown agent")
	}
}

func TestRunDoctorFixRejectsCursor(t *testing.T) {
	err := Run([]string{"doctor", "--fix", "--agent", "cursor"}, io.Discard)
	if err == nil {
		t.Fatal("expected error for cursor under --fix")
	}
	if !strings.Contains(err.Error(), "not installable") {
		t.Errorf("error %q should mention not installable", err.Error())
	}
}

func TestRunFlowMissingCommand(t *testing.T) {
	err := Run([]string{"flow"}, io.Discard)
	if err == nil {
		t.Fatal("expected error for missing flow command")
	}
}

func TestRunFlowUnsupportedCommand(t *testing.T) {
	err := Run([]string{"flow", "garbage"}, io.Discard)
	if err == nil {
		t.Fatal("expected error for unsupported flow command")
	}
	if !strings.Contains(err.Error(), "unsupported flow command") {
		t.Errorf("error %q should mention unsupported flow command", err.Error())
	}
}

func TestRunFlowQuickMissingChangeName(t *testing.T) {
	err := Run([]string{"flow", "quick"}, io.Discard)
	if err == nil {
		t.Fatal("expected error for missing change name")
	}
}

func TestParseKnownAgentAcceptsAllSupported(t *testing.T) {
	cases := []string{"codex", "claude-code", "opencode", "cursor", "vscode", "gemini-cli"}
	for _, raw := range cases {
		if _, err := parseKnownAgent(raw); err != nil {
			t.Errorf("parseKnownAgent(%q) returned error: %v", raw, err)
		}
	}
}

func TestParseInstallableAgentRejectsCursor(t *testing.T) {
	if _, err := parseInstallableAgent("cursor"); err == nil {
		t.Error("expected cursor to be rejected by parseInstallableAgent")
	}
}
