package status

import (
	"os"
	"path/filepath"
	"runtime"

	"nea-ai/internal/agents"
	"nea-ai/internal/components"
	"nea-ai/internal/model"
	"nea-ai/internal/system"
)

type Report struct {
	Version         string              `json:"version"`
	OS              string              `json:"os"`
	Arch            string              `json:"arch"`
	Paths           system.Paths        `json:"paths"`
	NeaBrain        BinaryStatus        `json:"neabrain"`
	OpenSpec        OpenSpecStatus      `json:"openspec"`
	Flow            FlowStatus          `json:"flow"`
	Components      []components.Status `json:"components"`
	Agents          []agents.Detection  `json:"agents"`
	SupportedPhase  string              `json:"supported_phase"`
	RecommendedNext string              `json:"recommended_next"`
}

type BinaryStatus struct {
	Available bool   `json:"available"`
	Path      string `json:"path,omitempty"`
}

type OpenSpecStatus struct {
	Present       bool   `json:"present"`
	Path          string `json:"path"`
	ConfigPresent bool   `json:"config_present"`
	StatusPresent bool   `json:"status_present"`
}

type FlowStatus struct {
	SkillsPresent        bool   `json:"skills_present"`
	SkillsPath           string `json:"skills_path"`
	ProjectPromptPresent bool   `json:"project_prompt_present"`
	ProjectPromptPath    string `json:"project_prompt_path"`
}

func Build(version string) (Report, error) {
	return BuildForAgent(version, model.AgentCodex)
}

func BuildForAgent(version string, agent model.AgentID) (Report, error) {
	paths, err := system.ResolvePaths()
	if err != nil {
		return Report{}, err
	}
	if agent == "" {
		agent = model.AgentCodex
	}
	registry := components.DefaultRegistry()
	componentStatuses := registry.DetectAll(components.ContextFromPaths(paths, agent))
	brainStatus := componentByID(componentStatuses, model.ComponentBrain)
	flowStatus := componentByID(componentStatuses, model.ComponentFlow)
	openSpecPath := filepath.Join(paths.WorkDir, "openspec")
	configPath := filepath.Join(openSpecPath, "config.yaml")
	statusPath := filepath.Join(openSpecPath, "changes", ".status.yaml")
	_, openSpecErr := os.Stat(openSpecPath)
	_, configErr := os.Stat(configPath)
	_, statusErr := os.Stat(statusPath)
	skillsPath := flowStatus.Details["skills_path"]
	projectPromptPath := flowStatus.Details["project_prompt_path"]
	report := Report{
		Version: version,
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		Paths:   paths,
		NeaBrain: BinaryStatus{
			Available: brainStatus.Present,
			Path:      brainStatus.Path,
		},
		OpenSpec: OpenSpecStatus{
			Present:       openSpecErr == nil,
			Path:          openSpecPath,
			ConfigPresent: configErr == nil,
			StatusPresent: statusErr == nil,
		},
		Flow: FlowStatus{
			SkillsPresent:        fileExists(skillsPath, "SKILL.md"),
			SkillsPath:           skillsPath,
			ProjectPromptPresent: exists(projectPromptPath),
			ProjectPromptPath:    projectPromptPath,
		},
		Components:      componentStatuses,
		Agents:          agents.DetectAll(paths.HomeDir),
		SupportedPhase:  "foundation",
		RecommendedNext: "Run `nea-ai init` in a target project, then `nea-ai doctor`.",
	}
	return report, nil
}

func componentByID(statuses []components.Status, id model.ComponentID) components.Status {
	for _, status := range statuses {
		if status.ID == id {
			return status
		}
	}
	return components.Status{ID: id, Details: map[string]string{}}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func fileExists(parts ...string) bool {
	return exists(filepath.Join(parts...))
}
