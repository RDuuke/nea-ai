package flowstate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	result.PendingTasks = parseInlineList(values["pending_tasks"])
	result.AwaitingApproval = parseBool(values["awaiting_approval"])
	result.Completed = parseBool(values["completed"])
	result.ModifiedArtifacts = parseInlineList(values["modified_artifacts"])
	result.Notes = values["notes"]
	result.NextRecommended = nextRecommended(result)
	return result, nil
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
