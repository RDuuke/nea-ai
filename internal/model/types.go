package model

type AgentID string

const (
	AgentCodex      AgentID = "codex"
	AgentClaudeCode AgentID = "claude-code"
	AgentOpenCode   AgentID = "opencode"
	AgentCursor     AgentID = "cursor"
	AgentVSCode     AgentID = "vscode"
	AgentGeminiCLI  AgentID = "gemini-cli"
)

type ComponentID string

const (
	ComponentBrain       ComponentID = "brain"
	ComponentFlow        ComponentID = "flow"
	ComponentMCP         ComponentID = "mcp"
	ComponentOpenSpec    ComponentID = "openspec"
	ComponentPermissions ComponentID = "permissions"
)

// SupportedAgents lists every agent the CLI recognises, including detection-only.
func SupportedAgents() []AgentID {
	return []AgentID{
		AgentCodex,
		AgentClaudeCode,
		AgentOpenCode,
		AgentCursor,
		AgentVSCode,
		AgentGeminiCLI,
	}
}

// InstallableAgents returns the agents that install/uninstall pipelines support today.
func InstallableAgents() []AgentID {
	return []AgentID{AgentCodex, AgentClaudeCode, AgentOpenCode}
}

// IsAgentKnown reports whether agent is part of SupportedAgents.
func IsAgentKnown(agent AgentID) bool {
	for _, candidate := range SupportedAgents() {
		if candidate == agent {
			return true
		}
	}
	return false
}

// IsAgentInstallable reports whether agent can be configured by install/uninstall.
func IsAgentInstallable(agent AgentID) bool {
	for _, candidate := range InstallableAgents() {
		if candidate == agent {
			return true
		}
	}
	return false
}

func SupportedComponents() []ComponentID {
	return []ComponentID{
		ComponentBrain,
		ComponentFlow,
		ComponentMCP,
		ComponentOpenSpec,
		ComponentPermissions,
	}
}
