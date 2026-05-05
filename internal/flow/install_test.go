package flow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nea-ai/internal/model"
)

func TestAgentLayoutOpenCode(t *testing.T) {
	layout := agentLayout("home", "work", model.AgentOpenCode)

	if layout.skillsTarget != filepath.Join("home", ".config", "opencode", "skills") {
		t.Fatalf("unexpected skills target: %s", layout.skillsTarget)
	}
	if layout.promptSource != filepath.Join("examples", "opencode", "AGENTS.md") {
		t.Fatalf("unexpected prompt source: %s", layout.promptSource)
	}
	if layout.commandsTarget != filepath.Join("home", ".config", "opencode", "commands") {
		t.Fatalf("unexpected commands target: %s", layout.commandsTarget)
	}
}

func TestUninstallProjectPromptRemovesMarkedBlockOnly(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "AGENTS.md")
	content := "before\n" + markerStart + "\nmanaged\n" + markerEnd + "\nafter\n"
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	backupPath, err := uninstallProjectPrompt(target)
	if err != nil {
		t.Fatal(err)
	}
	if backupPath == "" {
		t.Fatal("expected backup path")
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Contains(text, markerStart) || strings.Contains(text, "managed") {
		t.Fatalf("managed block still present: %q", text)
	}
	if !strings.Contains(text, "before") || !strings.Contains(text, "after") {
		t.Fatalf("unmanaged content missing: %q", text)
	}
}

func TestUninstallOpenCodeConfigRemovesFlowAgentsOnly(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "config.json")
	content := `{"agent":{"flow-nea-orchestrator":{},"custom-agent":{}}}`
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	changed, err := uninstallOpenCodeConfig(target)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected config to change")
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Contains(text, "flow-nea-orchestrator") {
		t.Fatalf("flow agent still present: %s", text)
	}
	if !strings.Contains(text, "custom-agent") {
		t.Fatalf("custom agent removed: %s", text)
	}
}

func TestAgentLayoutClaudeCode(t *testing.T) {
	layout := agentLayout("home", "work", model.AgentClaudeCode)

	if layout.skillsTarget != filepath.Join("home", ".claude", "skills") {
		t.Fatalf("unexpected skills target: %s", layout.skillsTarget)
	}
	if layout.promptSource != filepath.Join("examples", "claude-code", "CLAUDE.md") {
		t.Fatalf("unexpected prompt source: %s", layout.promptSource)
	}
	if layout.commandsTarget != filepath.Join("home", ".claude", "commands") {
		t.Fatalf("unexpected commands target: %s", layout.commandsTarget)
	}
}
