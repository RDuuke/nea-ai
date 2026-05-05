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

func SupportedComponents() []ComponentID {
	return []ComponentID{
		ComponentBrain,
		ComponentFlow,
		ComponentMCP,
		ComponentOpenSpec,
		ComponentPermissions,
	}
}
