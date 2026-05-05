package flow

import (
	"path/filepath"
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
