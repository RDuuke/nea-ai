package doctor

import (
	"os"

	"nea-ai/internal/components"
	"nea-ai/internal/model"
	"nea-ai/internal/status"
)

type CheckStatus string

const (
	CheckOK      CheckStatus = "ok"
	CheckWarning CheckStatus = "warning"
	CheckFailed  CheckStatus = "failed"
)

type Check struct {
	ID      string      `json:"id"`
	Status  CheckStatus `json:"status"`
	Message string      `json:"message"`
}

type Report struct {
	Ready  bool    `json:"ready"`
	Checks []Check `json:"checks"`
}

type FixReport struct {
	Agent      model.AgentID             `json:"agent"`
	Ready      bool                      `json:"ready"`
	Fixed      []model.ComponentID       `json:"fixed"`
	Skipped    []model.ComponentID       `json:"skipped,omitempty"`
	Components map[model.ComponentID]any `json:"components,omitempty"`
	Before     Report                    `json:"before"`
	After      Report                    `json:"after"`
}

func Run(version string) (Report, error) {
	return RunForAgent(version, model.AgentCodex)
}

func RunForAgent(version string, agent model.AgentID) (Report, error) {
	st, err := status.BuildForAgent(version, agent)
	if err != nil {
		return Report{}, err
	}
	componentChecks := components.DefaultRegistry().Checks(components.ContextFromPaths(st.Paths, agent))
	checks := []Check{
		checkBool("openspec.present", st.OpenSpec.Present, "openspec directory present", "openspec directory missing; run `nea-ai init`"),
		checkBool("openspec.config", st.OpenSpec.ConfigPresent, "openspec config present", "openspec/config.yaml missing"),
		checkWritable("workspace.writable", st.Paths.WorkDir),
	}
	for _, check := range componentChecks {
		checks = append(checks, Check{
			ID:      check.ID,
			Status:  CheckStatus(check.Status),
			Message: check.Message,
		})
	}
	ready := true
	for _, check := range checks {
		if check.Status == CheckFailed {
			ready = false
			break
		}
	}
	return Report{Ready: ready, Checks: checks}, nil
}

func FixForAgent(version string, agent model.AgentID) (FixReport, error) {
	if agent == "" {
		agent = model.AgentCodex
	}
	before, err := RunForAgent(version, agent)
	if err != nil {
		return FixReport{}, err
	}
	st, err := status.BuildForAgent(version, agent)
	if err != nil {
		return FixReport{}, err
	}

	registry := components.DefaultRegistry()
	ctx := components.ContextFromPaths(st.Paths, agent)
	results := map[model.ComponentID]any{}
	fixed := []model.ComponentID{}
	skipped := []model.ComponentID{}

	for _, componentStatus := range st.Components {
		if componentStatus.Installed {
			skipped = append(skipped, componentStatus.ID)
			continue
		}
		component, ok := registry.Get(componentStatus.ID)
		if !ok {
			skipped = append(skipped, componentStatus.ID)
			continue
		}
		result, err := component.Install(ctx)
		if err != nil {
			return FixReport{}, err
		}
		results[componentStatus.ID] = result
		fixed = append(fixed, componentStatus.ID)
	}

	after, err := RunForAgent(version, agent)
	if err != nil {
		return FixReport{}, err
	}
	return FixReport{
		Agent:      agent,
		Ready:      after.Ready,
		Fixed:      fixed,
		Skipped:    skipped,
		Components: results,
		Before:     before,
		After:      after,
	}, nil
}

func checkBool(id string, ok bool, okMessage string, failedMessage string) Check {
	if ok {
		return Check{ID: id, Status: CheckOK, Message: okMessage}
	}
	return Check{ID: id, Status: CheckFailed, Message: failedMessage}
}

func checkWritable(id string, dir string) Check {
	info, err := os.Stat(dir)
	if err != nil {
		return Check{ID: id, Status: CheckFailed, Message: err.Error()}
	}
	if !info.IsDir() {
		return Check{ID: id, Status: CheckFailed, Message: "workspace path is not a directory"}
	}
	return Check{ID: id, Status: CheckOK, Message: "workspace directory reachable"}
}
