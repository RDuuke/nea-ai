package brain

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"nea-ai/internal/agents"
	"nea-ai/internal/model"
	"nea-ai/internal/system"
)

type InstallOptions struct {
	Agent      model.AgentID
	Component  model.ComponentID
	WorkingDir string
}

type InstallResult struct {
	Agent         model.AgentID     `json:"agent"`
	Component     model.ComponentID `json:"component"`
	NeaBrainPath  string            `json:"neabrain_path"`
	ConfigPath    string            `json:"config_path"`
	BackupPath    string            `json:"backup_path,omitempty"`
	CommandOutput string            `json:"command_output"`
}

type UninstallOptions struct {
	Agent      model.AgentID
	Component  model.ComponentID
	WorkingDir string
}

type UninstallResult struct {
	Agent         model.AgentID     `json:"agent"`
	Component     model.ComponentID `json:"component"`
	NeaBrainPath  string            `json:"neabrain_path"`
	ConfigPath    string            `json:"config_path"`
	BackupPath    string            `json:"backup_path,omitempty"`
	CommandOutput string            `json:"command_output"`
}

func Install(options InstallOptions) (InstallResult, error) {
	if options.Agent == "" {
		options.Agent = model.AgentCodex
	}
	if options.Component == "" {
		options.Component = model.ComponentBrain
	}
	if !supportedAgent(options.Agent) {
		return InstallResult{}, fmt.Errorf("brain install does not support agent %q", options.Agent)
	}
	if options.Component != model.ComponentBrain {
		return InstallResult{}, fmt.Errorf("unsupported component %q", options.Component)
	}

	paths, err := system.ResolvePaths()
	if err != nil {
		return InstallResult{}, err
	}
	if options.WorkingDir == "" {
		options.WorkingDir = paths.WorkDir
	}

	brainPath, err := ResolveNeaBrain(options.WorkingDir)
	if err != nil {
		return InstallResult{}, err
	}

	configPath := agents.ConfigPath(paths.HomeDir, options.Agent)
	backupPath, err := backupIfExists(configPath)
	if err != nil {
		return InstallResult{}, err
	}

	cmd := exec.Command(brainPath, "setup", string(options.Agent), "--install")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return InstallResult{}, fmt.Errorf("run neabrain setup: %w\n%s", err, strings.TrimSpace(string(output)))
	}

	return InstallResult{
		Agent:         options.Agent,
		Component:     options.Component,
		NeaBrainPath:  brainPath,
		ConfigPath:    configPath,
		BackupPath:    backupPath,
		CommandOutput: strings.TrimSpace(string(output)),
	}, nil
}

func Uninstall(options UninstallOptions) (UninstallResult, error) {
	if options.Agent == "" {
		options.Agent = model.AgentCodex
	}
	if options.Component == "" {
		options.Component = model.ComponentBrain
	}
	if !supportedAgent(options.Agent) {
		return UninstallResult{}, fmt.Errorf("brain uninstall does not support agent %q", options.Agent)
	}
	if options.Component != model.ComponentBrain {
		return UninstallResult{}, fmt.Errorf("unsupported component %q", options.Component)
	}

	paths, err := system.ResolvePaths()
	if err != nil {
		return UninstallResult{}, err
	}
	if options.WorkingDir == "" {
		options.WorkingDir = paths.WorkDir
	}

	brainPath, err := ResolveNeaBrain(options.WorkingDir)
	if err != nil {
		return UninstallResult{}, err
	}

	configPath := agents.ConfigPath(paths.HomeDir, options.Agent)
	backupPath, err := backupIfExists(configPath)
	if err != nil {
		return UninstallResult{}, err
	}

	cmd := exec.Command(brainPath, "setup", string(options.Agent), "--uninstall")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return UninstallResult{}, fmt.Errorf("run neabrain setup uninstall: %w\n%s", err, strings.TrimSpace(string(output)))
	}

	return UninstallResult{
		Agent:         options.Agent,
		Component:     options.Component,
		NeaBrainPath:  brainPath,
		ConfigPath:    configPath,
		BackupPath:    backupPath,
		CommandOutput: strings.TrimSpace(string(output)),
	}, nil
}

func supportedAgent(agent model.AgentID) bool {
	switch agent {
	case model.AgentCodex, model.AgentOpenCode, model.AgentClaudeCode:
		return true
	default:
		return false
	}
}

func ResolveNeaBrain(workDir string) (string, error) {
	if path, err := exec.LookPath("neabrain"); err == nil {
		return path, nil
	}

	candidates := siblingCandidates(workDir)
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("neabrain binary not found in PATH or sibling repo; build or install https://github.com/RDuuke/nea-brain")
}

func siblingCandidates(workDir string) []string {
	parent := filepath.Dir(workDir)
	name := "neabrain"
	if runtime.GOOS == "windows" {
		name = "neabrain.exe"
	}
	return []string{
		filepath.Join(parent, "nea-brain", name),
		filepath.Join(parent, "neabrain", name),
	}
}

func backupIfExists(path string) (string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	backupPath := fmt.Sprintf("%s.nea-ai.%s.bak", path, time.Now().UTC().Format("20060102150405"))
	if err := os.MkdirAll(filepath.Dir(backupPath), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(backupPath, data, 0o644); err != nil {
		return "", err
	}
	return backupPath, nil
}
