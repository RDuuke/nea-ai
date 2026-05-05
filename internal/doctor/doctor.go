package doctor

import (
	"os"

	"nea-ai/internal/components"
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

func Run(version string) (Report, error) {
	st, err := status.Build(version)
	if err != nil {
		return Report{}, err
	}
	componentChecks := components.DefaultRegistry().Checks(components.ContextFromPaths(st.Paths, ""))
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
