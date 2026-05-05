package flow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

func Install(options InstallOptions) (InstallResult, error) {
	if options.Agent == "" {
		options.Agent = model.AgentCodex
	}
	if options.Agent != model.AgentCodex {
		return InstallResult{}, fmt.Errorf("flow install currently supports only agent %q", model.AgentCodex)
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

	skillsSource := filepath.Join(sourceRepo, "skills")
	skillsTarget := filepath.Join(paths.HomeDir, ".codex", "skills")
	filesCopied, err := copyTree(skillsSource, skillsTarget)
	if err != nil {
		return InstallResult{}, err
	}

	promptSource := filepath.Join(sourceRepo, "examples", "codex", "agents.md")
	promptTarget := filepath.Join(options.WorkingDir, "AGENTS.md")
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
	return "", fmt.Errorf("tdd-nea-flow repo not found next to %s", workDir)
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
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return err
		}
		count++
		return nil
	})
	return count, err
}

func installProjectPrompt(source string, target string) (string, error) {
	data, err := os.ReadFile(source)
	if err != nil {
		return "", err
	}
	block := markerStart + "\n" + string(data) + "\n" + markerEnd + "\n"

	existing, err := os.ReadFile(target)
	if os.IsNotExist(err) {
		return "", os.WriteFile(target, []byte(block), 0o644)
	}
	if err != nil {
		return "", err
	}
	text := string(existing)
	backupPath, err := backupFile(target, existing)
	if err != nil {
		return "", err
	}

	if strings.Contains(text, markerStart) && strings.Contains(text, markerEnd) {
		next := replaceMarkedBlock(text, block)
		return backupPath, os.WriteFile(target, []byte(next), 0o644)
	}

	if strings.TrimSpace(text) != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	text += "\n" + block
	return backupPath, os.WriteFile(target, []byte(text), 0o644)
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

func backupFile(path string, data []byte) (string, error) {
	backupPath := fmt.Sprintf("%s.nea-ai.%s.bak", path, time.Now().UTC().Format("20060102150405"))
	if err := os.WriteFile(backupPath, data, 0o644); err != nil {
		return "", err
	}
	return backupPath, nil
}
