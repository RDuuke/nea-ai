package agents

import (
	"os"
	"os/exec"
	"path/filepath"

	"nea-ai/internal/model"
)

type Detection struct {
	ID             model.AgentID `json:"id"`
	Installed      bool          `json:"installed"`
	ConfigPath     string        `json:"config_path,omitempty"`
	ConfigPresent  bool          `json:"config_present"`
	ExecutablePath string        `json:"executable_path,omitempty"`
}

func DetectAll(homeDir string) []Detection {
	agents := model.SupportedAgents()
	out := make([]Detection, 0, len(agents))
	for _, id := range agents {
		out = append(out, Detect(homeDir, id))
	}
	return out
}

func Detect(homeDir string, id model.AgentID) Detection {
	configPath := ConfigPath(homeDir, id)
	executableName := ExecutableName(id)
	executablePath, _ := exec.LookPath(executableName)
	_, statErr := os.Stat(configPath)
	return Detection{
		ID:             id,
		Installed:      executablePath != "" || statErr == nil,
		ConfigPath:     configPath,
		ConfigPresent:  statErr == nil,
		ExecutablePath: executablePath,
	}
}

func ConfigPath(homeDir string, id model.AgentID) string {
	switch id {
	case model.AgentCodex:
		return filepath.Join(homeDir, ".codex", "config.toml")
	case model.AgentClaudeCode:
		return filepath.Join(homeDir, ".claude", "settings.json")
	case model.AgentOpenCode:
		return filepath.Join(homeDir, ".config", "opencode", "config.json")
	case model.AgentCursor:
		return filepath.Join(homeDir, ".cursor", "mcp.json")
	case model.AgentVSCode:
		return filepath.Join(homeDir, "AppData", "Roaming", "Code", "User", "mcp.json")
	case model.AgentGeminiCLI:
		return filepath.Join(homeDir, ".gemini", "settings.json")
	default:
		return ""
	}
}

func ExecutableName(id model.AgentID) string {
	switch id {
	case model.AgentCodex:
		return "codex"
	case model.AgentClaudeCode:
		return "claude"
	case model.AgentOpenCode:
		return "opencode"
	case model.AgentCursor:
		return "cursor"
	case model.AgentVSCode:
		return "code"
	case model.AgentGeminiCLI:
		return "gemini"
	default:
		return string(id)
	}
}
