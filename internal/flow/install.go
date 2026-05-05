package flow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nea-ai/internal/model"
	"nea-ai/internal/system"
)

const markerStart = "<!-- nea-ai:flow-codex:start -->"
const markerEnd = "<!-- nea-ai:flow-codex:end -->"

type InstallOptions struct {
	Agent      model.AgentID
	WorkingDir string
}

type InstallResult struct {
	Agent            model.AgentID     `json:"agent"`
	Component        model.ComponentID `json:"component"`
	SourceRepo       string            `json:"source_repo"`
	SkillsSource     string            `json:"skills_source"`
	SkillsTarget     string            `json:"skills_target"`
	ProjectPrompt    string            `json:"project_prompt"`
	ProjectPromptBak string            `json:"project_prompt_backup,omitempty"`
	FilesCopied      int               `json:"files_copied"`
}

type UninstallOptions struct {
	Agent      model.AgentID
	WorkingDir string
}

type UninstallResult struct {
	Agent            model.AgentID     `json:"agent"`
	Component        model.ComponentID `json:"component"`
	SkillsTarget     string            `json:"skills_target"`
	ProjectPrompt    string            `json:"project_prompt"`
	ProjectPromptBak string            `json:"project_prompt_backup,omitempty"`
	FilesRemoved     int               `json:"files_removed"`
}

func Install(options InstallOptions) (InstallResult, error) {
	if options.Agent == "" {
		options.Agent = model.AgentCodex
	}
	if !model.IsAgentInstallable(options.Agent) {
		return InstallResult{}, fmt.Errorf("flow install does not support agent %q", options.Agent)
	}

	paths, err := system.ResolvePaths()
	if err != nil {
		return InstallResult{}, err
	}
	if options.WorkingDir == "" {
		options.WorkingDir = paths.WorkDir
	}

	sourceRepo, err := ResolveFlowRepo(options.WorkingDir)
	if err != nil {
		return InstallResult{}, err
	}

	layout := agentLayout(paths.HomeDir, options.WorkingDir, options.Agent)
	skillsSource := filepath.Join(sourceRepo, "skills")
	skillsTarget := layout.skillsTarget
	filesCopied, err := copyTree(skillsSource, skillsTarget)
	if err != nil {
		return InstallResult{}, err
	}

	if layout.commandsSource != "" && layout.commandsTarget != "" {
		if _, err := copyMarkdownFiles(filepath.Join(sourceRepo, layout.commandsSource), layout.commandsTarget); err != nil {
			return InstallResult{}, err
		}
	}
	if options.Agent == model.AgentOpenCode {
		if err := installOpenCodeConfig(filepath.Join(sourceRepo, "examples", "opencode", "opencode.multi.json"), filepath.Join(paths.HomeDir, ".config", "opencode", "config.json")); err != nil {
			return InstallResult{}, err
		}
	}

	promptSource := filepath.Join(sourceRepo, layout.promptSource)
	promptTarget := layout.promptTarget
	backupPath, err := installProjectPrompt(promptSource, promptTarget)
	if err != nil {
		return InstallResult{}, err
	}

	return InstallResult{
		Agent:            options.Agent,
		Component:        model.ComponentFlow,
		SourceRepo:       sourceRepo,
		SkillsSource:     skillsSource,
		SkillsTarget:     skillsTarget,
		ProjectPrompt:    promptTarget,
		ProjectPromptBak: backupPath,
		FilesCopied:      filesCopied,
	}, nil
}

func Uninstall(options UninstallOptions) (UninstallResult, error) {
	if options.Agent == "" {
		options.Agent = model.AgentCodex
	}
	if !model.IsAgentInstallable(options.Agent) {
		return UninstallResult{}, fmt.Errorf("flow uninstall does not support agent %q", options.Agent)
	}

	paths, err := system.ResolvePaths()
	if err != nil {
		return UninstallResult{}, err
	}
	if options.WorkingDir == "" {
		options.WorkingDir = paths.WorkDir
	}

	layout := agentLayout(paths.HomeDir, options.WorkingDir, options.Agent)
	removed, err := removeKnownFlowFiles(layout.skillsTarget, layout.commandsTarget)
	if err != nil {
		return UninstallResult{}, err
	}
	if options.Agent == model.AgentOpenCode {
		if changed, err := uninstallOpenCodeConfig(filepath.Join(paths.HomeDir, ".config", "opencode", "config.json")); err != nil {
			return UninstallResult{}, err
		} else if changed {
			removed++
		}
	}

	backupPath, err := uninstallProjectPrompt(layout.promptTarget)
	if err != nil {
		return UninstallResult{}, err
	}
	if backupPath != "" {
		removed++
	}

	return UninstallResult{
		Agent:            options.Agent,
		Component:        model.ComponentFlow,
		SkillsTarget:     layout.skillsTarget,
		ProjectPrompt:    layout.promptTarget,
		ProjectPromptBak: backupPath,
		FilesRemoved:     removed,
	}, nil
}

type layout struct {
	skillsTarget   string
	promptSource   string
	promptTarget   string
	commandsSource string
	commandsTarget string
}

func agentLayout(homeDir string, workDir string, agent model.AgentID) layout {
	switch agent {
	case model.AgentOpenCode:
		opencodeDir := filepath.Join(homeDir, ".config", "opencode")
		return layout{
			skillsTarget:   filepath.Join(opencodeDir, "skills"),
			promptSource:   filepath.Join("examples", "opencode", "AGENTS.md"),
			promptTarget:   filepath.Join(opencodeDir, "AGENTS.md"),
			commandsSource: filepath.Join("examples", "opencode", "commands"),
			commandsTarget: filepath.Join(opencodeDir, "commands"),
		}
	case model.AgentClaudeCode:
		claudeDir := filepath.Join(homeDir, ".claude")
		return layout{
			skillsTarget:   filepath.Join(claudeDir, "skills"),
			promptSource:   filepath.Join("examples", "claude-code", "CLAUDE.md"),
			promptTarget:   filepath.Join(claudeDir, "CLAUDE.md"),
			commandsSource: filepath.Join("examples", "claude-code", "commands"),
			commandsTarget: filepath.Join(claudeDir, "commands"),
		}
	default:
		return layout{
			skillsTarget: filepath.Join(homeDir, ".codex", "skills"),
			promptSource: filepath.Join("examples", "codex", "agents.md"),
			promptTarget: filepath.Join(workDir, "AGENTS.md"),
		}
	}
}

func ResolveFlowRepo(workDir string) (string, error) {
	parent := filepath.Dir(workDir)
	candidates := []string{
		filepath.Join(parent, "tdd-nea-flow"),
		filepath.Join(parent, "sdd-nea-flow"),
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(filepath.Join(candidate, "skills")); err == nil && info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("sdd-nea-flow repo not found next to %s; clone https://github.com/RDuuke/sdd-nea-flow", workDir)
}

func copyTree(source string, target string) (int, error) {
	count := 0
	err := filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(target, rel)
		if entry.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dest, data, 0o600); err != nil {
			return err
		}
		count++
		return nil
	})
	return count, err
}

func copyMarkdownFiles(source string, target string) (int, error) {
	entries, err := os.ReadDir(source)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		return 0, err
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(source, entry.Name()))
		if err != nil {
			return count, err
		}
		if err := os.WriteFile(filepath.Join(target, entry.Name()), data, 0o600); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func removeKnownFlowFiles(skillsTarget string, commandsTarget string) (int, error) {
	removed := 0
	for _, name := range []string{
		"flow-nea-apply",
		"flow-nea-archive",
		"flow-nea-continue",
		"flow-nea-design",
		"flow-nea-explore",
		"flow-nea-init",
		"flow-nea-propose",
		"flow-nea-quick",
		"flow-nea-spec",
		"flow-nea-tasks",
		"flow-nea-verify",
		"judgment-day",
		"skill-creator",
		"skill-registry",
		"_shared",
	} {
		if ok, err := removeIfExists(filepath.Join(skillsTarget, name)); err != nil {
			return removed, err
		} else if ok {
			removed++
		}
	}
	if commandsTarget != "" {
		for _, name := range []string{
			"flow-nea-apply.md",
			"flow-nea-archive.md",
			"flow-nea-continue.md",
			"flow-nea-design.md",
			"flow-nea-explore.md",
			"flow-nea-ff.md",
			"flow-nea-fix.md",
			"flow-nea-init.md",
			"flow-nea-judgment.md",
			"flow-nea-propose.md",
			"flow-nea-quick.md",
			"flow-nea-spec.md",
			"flow-nea-tasks.md",
			"flow-nea-verify.md",
			"skill-registry.md",
		} {
			if ok, err := removeIfExists(filepath.Join(commandsTarget, name)); err != nil {
				return removed, err
			} else if ok {
				removed++
			}
		}
	}
	return removed, nil
}

func removeIfExists(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	}
	return true, os.RemoveAll(path)
}

func installOpenCodeConfig(source string, target string) error {
	sourceData, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	var sourceConfig map[string]any
	if err := json.Unmarshal(sourceData, &sourceConfig); err != nil {
		return err
	}

	targetData, err := os.ReadFile(target)
	if os.IsNotExist(err) {
		targetData = []byte("{}")
	} else if err != nil {
		return err
	}
	var targetConfig map[string]any
	if err := json.Unmarshal(targetData, &targetConfig); err != nil {
		targetConfig = map[string]any{}
	}

	sourceAgents, _ := sourceConfig["agent"].(map[string]any)
	if len(sourceAgents) == 0 {
		return nil
	}
	targetAgents := getOrCreate(targetConfig, "agent")
	for name, value := range sourceAgents {
		targetAgents[name] = value
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(targetConfig, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(target, append(data, '\n'), 0o600)
}

func uninstallOpenCodeConfig(target string) (bool, error) {
	data, err := os.ReadFile(target)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return false, nil
	}
	agents, ok := config["agent"].(map[string]any)
	if !ok {
		return false, nil
	}
	changed := false
	for _, name := range []string{
		"flow-nea-orchestrator",
		"flow-nea-explore",
		"flow-nea-propose",
		"flow-nea-spec",
		"flow-nea-design",
		"flow-nea-tasks",
		"flow-nea-apply",
		"flow-nea-verify",
		"flow-nea-archive",
		"judgment-day-a",
		"judgment-day-b",
		"default",
	} {
		if _, ok := agents[name]; ok {
			delete(agents, name)
			changed = true
		}
	}
	if !changed {
		return false, nil
	}
	out, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return false, err
	}
	return true, os.WriteFile(target, append(out, '\n'), 0o600)
}

func getOrCreate(values map[string]any, key string) map[string]any {
	if child, ok := values[key].(map[string]any); ok {
		return child
	}
	child := map[string]any{}
	values[key] = child
	return child
}

func installProjectPrompt(source string, target string) (string, error) {
	data, err := os.ReadFile(source)
	if err != nil {
		return "", err
	}
	block := markerStart + "\n" + string(data) + "\n" + markerEnd + "\n"

	existing, err := os.ReadFile(target)
	if os.IsNotExist(err) {
		return "", os.WriteFile(target, []byte(block), 0o600)
	}
	if err != nil {
		return "", err
	}
	text := string(existing)
	backupPath, err := system.BackupFile(target)
	if err != nil {
		return "", err
	}

	if strings.Contains(text, markerStart) && strings.Contains(text, markerEnd) {
		next := replaceMarkedBlock(text, block)
		return backupPath, os.WriteFile(target, []byte(next), 0o600)
	}

	if strings.TrimSpace(text) != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	text += "\n" + block
	return backupPath, os.WriteFile(target, []byte(text), 0o600)
}

func uninstallProjectPrompt(target string) (string, error) {
	data, err := os.ReadFile(target)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	text := string(data)
	if !strings.Contains(text, markerStart) || !strings.Contains(text, markerEnd) {
		return "", nil
	}
	backupPath, err := system.BackupFile(target)
	if err != nil {
		return "", err
	}
	next := replaceMarkedBlock(text, "")
	return backupPath, os.WriteFile(target, []byte(next), 0o600)
}

func replaceMarkedBlock(text string, block string) string {
	start := strings.Index(text, markerStart)
	end := strings.Index(text, markerEnd)
	if start < 0 || end < 0 || end < start {
		return text
	}
	end += len(markerEnd)
	next := text[:start] + strings.TrimRight(block, "\n") + text[end:]
	if !strings.HasSuffix(next, "\n") {
		next += "\n"
	}
	return next
}
