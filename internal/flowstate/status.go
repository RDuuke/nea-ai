package flowstate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Status struct {
	OpenSpecPath      string   `json:"openspec_path"`
	StatusPath        string   `json:"status_path"`
	ConfigPath        string   `json:"config_path"`
	Present           bool     `json:"present"`
	ConfigPresent     bool     `json:"config_present"`
	StatusPresent     bool     `json:"status_present"`
	Change            string   `json:"change"`
	CurrentPhase      string   `json:"current_phase"`
	PendingTasks      []string `json:"pending_tasks"`
	AwaitingApproval  bool     `json:"awaiting_approval"`
	Completed         bool     `json:"completed"`
	ModifiedArtifacts []string `json:"modified_artifacts"`
	Notes             string   `json:"notes,omitempty"`
	NextRecommended   string   `json:"next_recommended"`
}

type QuickOptions struct {
	Name         string
	Title        string
	Objective    string
	Files        []string
	Blueprint    []string
	Risks        []string
	Verification []string
}

type QuickResult struct {
	Status           string   `json:"status"`
	ExecutiveSummary string   `json:"executive_summary"`
	Artifact         string   `json:"artifact"`
	NextRecommended  string   `json:"next_recommended"`
	Risks            []string `json:"risks"`
	SkillResolution  string   `json:"skill_resolution"`
}

var changeNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,48}[a-z0-9]$`)

func Build(workDir string) (Status, error) {
	openSpecPath := filepath.Join(workDir, "openspec")
	statusPath := filepath.Join(openSpecPath, "changes", ".status.yaml")
	configPath := filepath.Join(openSpecPath, "config.yaml")
	result := Status{
		OpenSpecPath: openSpecPath,
		StatusPath:   statusPath,
		ConfigPath:   configPath,
	}

	result.Present = exists(openSpecPath)
	result.ConfigPresent = exists(configPath)
	result.StatusPresent = exists(statusPath)
	if !result.StatusPresent {
		result.NextRecommended = "Run `nea-ai init` to create OpenSpec state."
		return result, nil
	}

	data, err := os.ReadFile(statusPath)
	if err != nil {
		return result, err
	}
	values := parseStatusYAML(string(data))
	result.Change = values["change"]
	result.CurrentPhase = values["current_phase"]
	if result.CurrentPhase == "" {
		result.CurrentPhase = values["phase"]
	}
	result.PendingTasks = parseInlineList(values["pending_tasks"])
	result.AwaitingApproval = parseBool(values["awaiting_approval"])
	result.Completed = parseBool(values["completed"])
	result.ModifiedArtifacts = parseInlineList(values["modified_artifacts"])
	result.Notes = values["notes"]
	result.NextRecommended = nextRecommended(result)
	return result, nil
}

func Quick(workDir string, options QuickOptions) (QuickResult, error) {
	if !changeNamePattern.MatchString(options.Name) {
		return QuickResult{}, fmt.Errorf("invalid change name %q; use kebab-case, 3-50 chars", options.Name)
	}
	current, err := Build(workDir)
	if err != nil {
		return QuickResult{}, err
	}
	if !current.Present || !current.ConfigPresent || !current.StatusPresent {
		return QuickResult{}, fmt.Errorf("openspec state missing; run `nea-ai init` first")
	}
	if current.AwaitingApproval {
		return QuickResult{}, fmt.Errorf("current change %q is awaiting approval", current.Change)
	}
	if current.Change != "" && !current.Completed {
		return QuickResult{}, fmt.Errorf("current change %q is still active", current.Change)
	}

	changeDir := filepath.Join(workDir, "openspec", "changes", options.Name)
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		return QuickResult{}, err
	}
	quickPath := filepath.Join(changeDir, "quick.md")
	if exists(quickPath) {
		return QuickResult{}, fmt.Errorf("quick artifact already exists: %s", quickPath)
	}
	if err := os.WriteFile(quickPath, []byte(renderQuick(options)), 0o600); err != nil {
		return QuickResult{}, err
	}

	statusPath := filepath.Join(workDir, "openspec", "changes", ".status.yaml")
	if err := os.WriteFile(statusPath, []byte(renderQuickStatus(options.Name)), 0o600); err != nil {
		return QuickResult{}, err
	}

	return QuickResult{
		Status:           "ok",
		ExecutiveSummary: "Quick blueprint creado y esperando aprobacion.",
		Artifact:         quickPath,
		NextRecommended:  "APPLY",
		Risks:            nonEmpty(options.Risks, []string{"Confirmar alcance antes de aplicar cambios."}),
		SkillResolution:  "native",
	}, nil
}

func renderQuick(options QuickOptions) string {
	title := options.Title
	if title == "" {
		title = strings.ReplaceAll(options.Name, "-", " ")
	}
	objective := options.Objective
	if objective == "" {
		objective = "Definir el resultado esperado antes de aplicar cambios."
	}
	files := nonEmpty(options.Files, []string{"Por definir"})
	blueprint := nonEmpty(options.Blueprint, []string{
		"Review the affected area with the minimum necessary context.",
		"Apply the change in a bounded way.",
		"Validate with the indicated commands.",
	})
	risks := nonEmpty(options.Risks, []string{"Scope may require the full flow if architecture or multi-domain changes appear."})
	verification := nonEmpty(options.Verification, []string{"go test ./...", "nea-ai flow status --json"})

	return fmt.Sprintf(`# Quick Fix: %s

## Objective

%s

## Affected Files

%s

## Blueprint

%s

## Risks

%s

## Verification

%s
`, title, objective, markdownList(files), markdownList(blueprint), markdownList(risks), markdownList(verification))
}

func renderQuickStatus(change string) string {
	return fmt.Sprintf(`change: "%s"
phase: QUICK
current_phase: QUICK
pending_tasks: []
awaiting_approval: true
completed: false
modified_artifacts: []
notes: "quick"
updated_at: "%s"
`, change, time.Now().UTC().Format(time.RFC3339))
}

func markdownList(items []string) string {
	var builder strings.Builder
	for _, item := range items {
		builder.WriteString("- ")
		builder.WriteString(item)
		builder.WriteString("\n")
	}
	return strings.TrimRight(builder.String(), "\n")
}

func nonEmpty(items []string, fallback []string) []string {
	var out []string
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}

func parseStatusYAML(data string) map[string]string {
	values := map[string]string{}
	for _, line := range strings.Split(data, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || !strings.Contains(trimmed, ":") {
			continue
		}
		parts := strings.SplitN(trimmed, ":", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		values[key] = strings.Trim(value, `"`)
	}
	return values
}

func parseInlineList(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" || value == "[]" {
		return []string{}
	}
	if !strings.HasPrefix(value, "[") || !strings.HasSuffix(value, "]") {
		return []string{strings.Trim(value, `"`)}
	}
	value = strings.TrimSuffix(strings.TrimPrefix(value, "["), "]")
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.Trim(strings.TrimSpace(part), `"`)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func parseBool(value string) bool {
	return strings.EqualFold(strings.TrimSpace(value), "true")
}

func nextRecommended(status Status) string {
	if !status.Present || !status.ConfigPresent || !status.StatusPresent {
		return "Run `nea-ai init` to create OpenSpec state."
	}
	if status.AwaitingApproval {
		return fmt.Sprintf("Approve or reject the current %s phase before continuing.", status.CurrentPhase)
	}
	if status.Completed {
		return "Flow is completed. Archive or start a new change."
	}
	if status.Change == "" {
		return "Start a change with `nea-ai flow quick <name>` or the Flow-NEA skills."
	}
	if len(status.PendingTasks) > 0 {
		return "Continue APPLY for pending tasks."
	}
	switch strings.ToUpper(status.CurrentPhase) {
	case "INIT":
		return "Start EXPLORE or QUICK for a new change."
	case "TASKS":
		return "Run APPLY for the current change."
	case "APPLY":
		return "Run VERIFY for the current change."
	case "VERIFY":
		return "Archive the verified change."
	default:
		return "Continue the Flow-NEA phase sequence."
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
