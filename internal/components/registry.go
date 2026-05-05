package components

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"nea-ai/internal/agents"
	"nea-ai/internal/brain"
	"nea-ai/internal/flow"
	"nea-ai/internal/model"
	"nea-ai/internal/system"
)

type InstallContext struct {
	Agent   model.AgentID
	WorkDir string
	HomeDir string
}

type Status struct {
	ID        model.ComponentID `json:"id"`
	Present   bool              `json:"present"`
	Installed bool              `json:"installed"`
	Path      string            `json:"path,omitempty"`
	Details   map[string]string `json:"details,omitempty"`
}

type CheckStatus string

const (
	CheckOK     CheckStatus = "ok"
	CheckFailed CheckStatus = "failed"
)

type Check struct {
	ID      string      `json:"id"`
	Status  CheckStatus `json:"status"`
	Message string      `json:"message"`
}

type Component interface {
	ID() model.ComponentID
	Install(InstallContext) (any, error)
	Detect(InstallContext) Status
	Checks(InstallContext) []Check
}

type Registry struct {
	components map[model.ComponentID]Component
	order      []model.ComponentID
}

func DefaultRegistry() Registry {
	items := []Component{
		brainComponent{},
		flowComponent{},
	}
	registry := Registry{
		components: make(map[model.ComponentID]Component, len(items)),
		order:      make([]model.ComponentID, 0, len(items)),
	}
	for _, item := range items {
		registry.components[item.ID()] = item
		registry.order = append(registry.order, item.ID())
	}
	return registry
}

func ContextFromPaths(paths system.Paths, agent model.AgentID) InstallContext {
	if agent == "" {
		agent = model.AgentCodex
	}
	return InstallContext{
		Agent:   agent,
		WorkDir: paths.WorkDir,
		HomeDir: paths.HomeDir,
	}
}

func (r Registry) Get(id model.ComponentID) (Component, bool) {
	component, ok := r.components[id]
	return component, ok
}

func (r Registry) DetectAll(ctx InstallContext) []Status {
	out := make([]Status, 0, len(r.order))
	for _, id := range r.order {
		out = append(out, r.components[id].Detect(ctx))
	}
	return out
}

func (r Registry) Checks(ctx InstallContext) []Check {
	var out []Check
	for _, id := range r.order {
		out = append(out, r.components[id].Checks(ctx)...)
	}
	return out
}

type brainComponent struct{}

func (brainComponent) ID() model.ComponentID {
	return model.ComponentBrain
}

func (brainComponent) Install(ctx InstallContext) (any, error) {
	result, err := brain.Install(brain.InstallOptions{
		Agent:      ctx.Agent,
		Component:  model.ComponentBrain,
		WorkingDir: ctx.WorkDir,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (brainComponent) Detect(ctx InstallContext) Status {
	brainPath, _ := brain.ResolveNeaBrain(ctx.WorkDir)
	configPath := agents.ConfigPath(ctx.HomeDir, ctx.Agent)
	configured := brainConfigured(configPath, ctx.Agent)
	return Status{
		ID:        model.ComponentBrain,
		Present:   brainPath != "",
		Installed: brainPath != "" && configured,
		Path:      brainPath,
		Details: map[string]string{
			"config_path": configPath,
			"agent":       string(ctx.Agent),
		},
	}
}

func (component brainComponent) Checks(ctx InstallContext) []Check {
	status := component.Detect(ctx)
	agentName := string(ctx.Agent)
	return []Check{
		checkBool("neabrain.binary", status.Present, "neabrain binary resolved", "neabrain binary not found in PATH or sibling repo"),
		checkBool(agentName+".neabrain_mcp", status.Installed, agentName+" NeaBrain MCP configured", agentName+" NeaBrain MCP missing; run `nea-ai install --agent "+agentName+" --components brain`"),
	}
}

type flowComponent struct{}

func (flowComponent) ID() model.ComponentID {
	return model.ComponentFlow
}

func (flowComponent) Install(ctx InstallContext) (any, error) {
	result, err := flow.Install(flow.InstallOptions{
		Agent:      ctx.Agent,
		WorkingDir: ctx.WorkDir,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (flowComponent) Detect(ctx InstallContext) Status {
	skillsPath := flowSkillsPath(ctx.HomeDir, ctx.Agent)
	projectPromptPath := flowPromptPath(ctx.HomeDir, ctx.WorkDir, ctx.Agent)
	skillsPresent := exists(skillsPath)
	promptPresent := exists(projectPromptPath)
	return Status{
		ID:        model.ComponentFlow,
		Present:   skillsPresent || promptPresent,
		Installed: skillsPresent && promptPresent,
		Details: map[string]string{
			"skills_path":         filepath.Dir(skillsPath),
			"project_prompt_path": projectPromptPath,
			"agent":               string(ctx.Agent),
		},
	}
}

func (component flowComponent) Checks(ctx InstallContext) []Check {
	status := component.Detect(ctx)
	agentName := string(ctx.Agent)
	return []Check{
		checkBool("flow.skills."+agentName, fileExists(status.Details["skills_path"], "SKILL.md"), "NEA Flow skills installed for "+agentName, "NEA Flow skills missing; run `nea-ai install --agent "+agentName+" --components flow`"),
		checkBool("flow.prompt."+agentName, exists(status.Details["project_prompt_path"]), "NEA Flow prompt present for "+agentName, "NEA Flow prompt missing; run `nea-ai install --agent "+agentName+" --components flow`"),
	}
}

func checkBool(id string, ok bool, okMessage string, failedMessage string) Check {
	if ok {
		return Check{ID: id, Status: CheckOK, Message: okMessage}
	}
	return Check{ID: id, Status: CheckFailed, Message: failedMessage}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func fileExists(parts ...string) bool {
	return exists(filepath.Join(parts...))
}

func fileContains(path string, needle string) bool {
	data, err := os.ReadFile(path)
	return err == nil && strings.Contains(string(data), needle)
}

func brainConfigured(configPath string, agent model.AgentID) bool {
	switch agent {
	case model.AgentCodex:
		return fileContains(configPath, "[mcp_servers.neabrain]")
	case model.AgentOpenCode:
		return jsonPathExists(configPath, "mcp", "servers", "neabrain")
	case model.AgentClaudeCode:
		return jsonPathExists(configPath, "mcpServers", "neabrain")
	default:
		return false
	}
}

func jsonPathExists(path string, keys ...string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return false
	}
	current := root
	for index, key := range keys {
		value, ok := current[key]
		if !ok {
			return false
		}
		if index == len(keys)-1 {
			return true
		}
		next, ok := value.(map[string]any)
		if !ok {
			return false
		}
		current = next
	}
	return false
}

func flowSkillsPath(homeDir string, agent model.AgentID) string {
	switch agent {
	case model.AgentOpenCode:
		return filepath.Join(homeDir, ".config", "opencode", "skills", "flow-nea-init", "SKILL.md")
	case model.AgentClaudeCode:
		return filepath.Join(homeDir, ".claude", "skills", "flow-nea-init", "SKILL.md")
	default:
		return filepath.Join(homeDir, ".codex", "skills", "flow-nea-init", "SKILL.md")
	}
}

func flowPromptPath(homeDir string, workDir string, agent model.AgentID) string {
	switch agent {
	case model.AgentOpenCode:
		return filepath.Join(homeDir, ".config", "opencode", "AGENTS.md")
	case model.AgentClaudeCode:
		return filepath.Join(homeDir, ".claude", "CLAUDE.md")
	default:
		return filepath.Join(workDir, "AGENTS.md")
	}
}
