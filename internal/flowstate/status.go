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

type PhaseOptions struct {
	Name      string
	Title     string
	Objective string
	Summary   string
	Files     []string
	Commands  []string
	Risks     []string
}

type PhaseResult struct {
	Status           string   `json:"status"`
	ExecutiveSummary string   `json:"executive_summary"`
	Artifacts        []string `json:"artifacts"`
	CurrentPhase     string   `json:"current_phase"`
	NextRecommended  string   `json:"next_recommended"`
	Risks            []string `json:"risks"`
	SkillResolution  string   `json:"skill_resolution"`
}

type ContinueResult struct {
	Status           string   `json:"status"`
	ExecutiveSummary string   `json:"executive_summary"`
	Change           string   `json:"change"`
	CurrentPhase     string   `json:"current_phase"`
	NextPhase        string   `json:"next_phase"`
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
	if err := validateChangeName(options.Name); err != nil {
		return QuickResult{}, err
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
		ExecutiveSummary: "Quick blueprint created and awaiting approval.",
		Artifact:         quickPath,
		NextRecommended:  "APPLY",
		Risks:            nonEmpty(options.Risks, []string{"Confirm scope before applying changes."}),
		SkillResolution:  "native",
	}, nil
}

func Explore(workDir string, options PhaseOptions) (PhaseResult, error) {
	if err := validateChangeName(options.Name); err != nil {
		return PhaseResult{}, err
	}
	current, err := Build(workDir)
	if err != nil {
		return PhaseResult{}, err
	}
	if err := requireOpenSpecReady(current); err != nil {
		return PhaseResult{}, err
	}
	if err := requireNoActiveChange(current); err != nil {
		return PhaseResult{}, err
	}

	changeDir := filepath.Join(workDir, "openspec", "changes", options.Name)
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		return PhaseResult{}, err
	}
	artifact := filepath.Join(changeDir, "exploration.md")
	if exists(artifact) {
		return PhaseResult{}, fmt.Errorf("exploration artifact already exists: %s", artifact)
	}
	if err := os.WriteFile(artifact, []byte(renderExplore(options)), 0o600); err != nil {
		return PhaseResult{}, err
	}
	if err := writeStatus(workDir, options.Name, "EXPLORE", false, false, []string{}, "explore"); err != nil {
		return PhaseResult{}, err
	}
	return PhaseResult{
		Status:           "ok",
		ExecutiveSummary: "Exploration artifact created.",
		Artifacts:        []string{artifact},
		CurrentPhase:     "EXPLORE",
		NextRecommended:  "Run `nea-ai flow propose " + options.Name + "` when the approach is clear.",
		Risks:            nonEmpty(options.Risks, []string{}),
		SkillResolution:  "native",
	}, nil
}

func Propose(workDir string, options PhaseOptions) (PhaseResult, error) {
	if err := validateChangeName(options.Name); err != nil {
		return PhaseResult{}, err
	}
	current, err := Build(workDir)
	if err != nil {
		return PhaseResult{}, err
	}
	if err := requireOpenSpecReady(current); err != nil {
		return PhaseResult{}, err
	}
	if current.AwaitingApproval {
		return PhaseResult{}, fmt.Errorf("current change %q is awaiting approval", current.Change)
	}
	if current.Change != "" && current.Change != options.Name && !current.Completed {
		return PhaseResult{}, fmt.Errorf("current change %q is still active", current.Change)
	}

	changeDir := filepath.Join(workDir, "openspec", "changes", options.Name)
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		return PhaseResult{}, err
	}
	artifact := filepath.Join(changeDir, "proposal.md")
	if exists(artifact) {
		return PhaseResult{}, fmt.Errorf("proposal artifact already exists: %s", artifact)
	}
	if err := os.WriteFile(artifact, []byte(renderProposal(options)), 0o600); err != nil {
		return PhaseResult{}, err
	}
	if err := writeStatus(workDir, options.Name, "PROPOSE", true, false, []string{}, "proposal awaiting approval"); err != nil {
		return PhaseResult{}, err
	}
	return PhaseResult{
		Status:           "ok",
		ExecutiveSummary: "Proposal artifact created and awaiting approval.",
		Artifacts:        []string{artifact},
		CurrentPhase:     "PROPOSE",
		NextRecommended:  "Approve the proposal, then continue to SPEC and DESIGN.",
		Risks:            nonEmpty(options.Risks, []string{"Confirm proposal scope before implementation planning."}),
		SkillResolution:  "native",
	}, nil
}

func Continue(workDir string) (ContinueResult, error) {
	current, err := Build(workDir)
	if err != nil {
		return ContinueResult{}, err
	}
	if err := requireOpenSpecReady(current); err != nil {
		return ContinueResult{}, err
	}
	if current.AwaitingApproval {
		return ContinueResult{}, fmt.Errorf("current change %q is awaiting approval", current.Change)
	}
	nextPhase := nextPhase(current)
	return ContinueResult{
		Status:           "ok",
		ExecutiveSummary: "Flow state inspected.",
		Change:           current.Change,
		CurrentPhase:     current.CurrentPhase,
		NextPhase:        nextPhase,
		NextRecommended:  nextCommand(current, nextPhase),
		Risks:            []string{},
		SkillResolution:  "native",
	}, nil
}

func Verify(workDir string, options PhaseOptions) (PhaseResult, error) {
	current, err := Build(workDir)
	if err != nil {
		return PhaseResult{}, err
	}
	if err := requireOpenSpecReady(current); err != nil {
		return PhaseResult{}, err
	}
	change := options.Name
	if change == "" {
		change = current.Change
	}
	if err := validateChangeName(change); err != nil {
		return PhaseResult{}, err
	}
	if current.Change != "" && current.Change != change && !current.Completed {
		return PhaseResult{}, fmt.Errorf("current change %q is still active", current.Change)
	}

	changeDir := filepath.Join(workDir, "openspec", "changes", change)
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		return PhaseResult{}, err
	}
	artifact := filepath.Join(changeDir, "verify-report.md")
	if err := os.WriteFile(artifact, []byte(renderVerify(change, options)), 0o600); err != nil {
		return PhaseResult{}, err
	}
	if err := writeStatus(workDir, change, "VERIFY", false, true, []string{}, "verified"); err != nil {
		return PhaseResult{}, err
	}
	return PhaseResult{
		Status:           "ok",
		ExecutiveSummary: "Verification report written.",
		Artifacts:        []string{artifact},
		CurrentPhase:     "VERIFY",
		NextRecommended:  "Archive the verified change when ready.",
		Risks:            nonEmpty(options.Risks, []string{}),
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
		objective = "Define the expected result before applying changes."
	}
	files := nonEmpty(options.Files, []string{"TBD"})
	blueprint := nonEmpty(options.Blueprint, []string{
		"Review the affected area with the minimum necessary context.",
		"Apply the change in a bounded way.",
		"Validate with the indicated commands.",
	})
	risks := nonEmpty(options.Risks, []string{"Scope may require the full flow if architecture or multi-domain changes appear."})
	verification := nonEmpty(options.Verification, []string{"go test ./...", "nea-ai flow status"})

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

func renderExplore(options PhaseOptions) string {
	title := defaultString(options.Title, strings.ReplaceAll(options.Name, "-", " "))
	objective := defaultString(options.Objective, "Understand the change before proposing implementation.")
	summary := defaultString(options.Summary, "Capture findings, constraints, and candidate approaches.")
	files := nonEmpty(options.Files, []string{"TBD"})

	return fmt.Sprintf(`# Exploration: %s

## Objective

%s

## Summary

%s

## Files

%s

## Risks

%s
`, title, objective, summary, markdownList(files), markdownList(nonEmpty(options.Risks, []string{"TBD"})))
}

func renderProposal(options PhaseOptions) string {
	title := defaultString(options.Title, strings.ReplaceAll(options.Name, "-", " "))
	objective := defaultString(options.Objective, "Define the intended change and scope.")
	summary := defaultString(options.Summary, "Describe approach, boundaries, and acceptance criteria.")

	return fmt.Sprintf(`# Proposal: %s

## Objective

%s

## Scope

%s

## Affected Files

%s

## Risks

%s
`, title, objective, summary, markdownList(nonEmpty(options.Files, []string{"TBD"})), markdownList(nonEmpty(options.Risks, []string{"TBD"})))
}

func renderVerify(change string, options PhaseOptions) string {
	summary := defaultString(options.Summary, "Verification completed.")
	commands := nonEmpty(options.Commands, []string{"go test ./..."})

	return fmt.Sprintf(`# Verify Report: %s

## Summary

%s

## Commands

%s

## Risks

%s
`, change, summary, markdownList(commands), markdownList(nonEmpty(options.Risks, []string{"None"})))
}

func renderQuickStatus(change string) string {
	return renderStatus(change, "QUICK", true, false, []string{}, "quick")
}

func writeStatus(workDir string, change string, phase string, awaitingApproval bool, completed bool, pendingTasks []string, notes string) error {
	statusPath := filepath.Join(workDir, "openspec", "changes", ".status.yaml")
	return os.WriteFile(statusPath, []byte(renderStatus(change, phase, awaitingApproval, completed, pendingTasks, notes)), 0o600)
}

func renderStatus(change string, phase string, awaitingApproval bool, completed bool, pendingTasks []string, notes string) string {
	return fmt.Sprintf(`change: "%s"
phase: %s
current_phase: %s
pending_tasks: %s
awaiting_approval: %t
completed: %t
modified_artifacts: []
notes: "%s"
updated_at: "%s"
`, change, phase, phase, formatInlineList(pendingTasks), awaitingApproval, completed, notes, time.Now().UTC().Format(time.RFC3339))
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

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func validateChangeName(name string) error {
	if !changeNamePattern.MatchString(name) {
		return fmt.Errorf("invalid change name %q; use kebab-case, 3-50 chars", name)
	}
	return nil
}

func requireOpenSpecReady(status Status) error {
	if !status.Present || !status.ConfigPresent || !status.StatusPresent {
		return fmt.Errorf("openspec state missing; run `nea-ai init` first")
	}
	return nil
}

func requireNoActiveChange(status Status) error {
	if status.AwaitingApproval {
		return fmt.Errorf("current change %q is awaiting approval", status.Change)
	}
	if status.Change != "" && !status.Completed {
		return fmt.Errorf("current change %q is still active", status.Change)
	}
	return nil
}

func formatInlineList(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	quoted := make([]string, 0, len(items))
	for _, item := range items {
		quoted = append(quoted, fmt.Sprintf("%q", item))
	}
	return "[" + strings.Join(quoted, ", ") + "]"
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

func nextPhase(status Status) string {
	if status.Completed {
		return "ARCHIVE"
	}
	switch strings.ToUpper(status.CurrentPhase) {
	case "INIT":
		return "EXPLORE"
	case "EXPLORE":
		return "PROPOSE"
	case "PROPOSE":
		return "SPEC"
	case "SPEC", "DESIGN":
		return "TASKS"
	case "TASKS":
		return "APPLY"
	case "APPLY":
		return "VERIFY"
	case "VERIFY":
		return "ARCHIVE"
	default:
		return "EXPLORE"
	}
}

func nextCommand(status Status, phase string) string {
	if status.Change == "" {
		return "Run `nea-ai flow explore <change-name>` or `nea-ai flow quick <change-name>`."
	}
	switch phase {
	case "PROPOSE":
		return "Run `nea-ai flow propose " + status.Change + "`."
	case "VERIFY":
		return "Run `nea-ai flow verify " + status.Change + "` after implementation checks pass."
	case "ARCHIVE":
		return "Archive the completed change."
	default:
		return "Continue with phase " + phase + " using Flow-NEA skills."
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
