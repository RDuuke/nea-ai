package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"nea-ai/internal/components"
	"nea-ai/internal/doctor"
	"nea-ai/internal/flowstate"
	"nea-ai/internal/model"
	"nea-ai/internal/openspec"
	"nea-ai/internal/status"
	"nea-ai/internal/system"
)

var Version = "dev"

func Run(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		printHelp(stdout)
		return nil
	}

	switch args[0] {
	case "version", "--version", "-v":
		_, _ = fmt.Fprintf(stdout, "nea-ai %s\n", Version)
		return nil
	case "help", "--help", "-h":
		printHelp(stdout)
		return nil
	case "status":
		report, err := runStatus(args[1:])
		if err != nil {
			return err
		}
		return writeJSON(stdout, report)
	case "doctor":
		report, err := runDoctorValue(args[1:])
		if err != nil {
			return err
		}
		return writeJSON(stdout, report)
	case "init":
		workDir, err := os.Getwd()
		if err != nil {
			return err
		}
		result, err := openspec.Init(workDir)
		if err != nil {
			return err
		}
		return writeJSON(stdout, result)
	case "flow":
		result, err := runFlow(args[1:])
		if err != nil {
			return err
		}
		return writeJSON(stdout, result)
	case "install":
		result, err := runInstall(args[1:])
		if err != nil {
			return err
		}
		return writeJSON(stdout, result)
	case "uninstall":
		result, err := runUninstall(args[1:])
		if err != nil {
			return err
		}
		return writeJSON(stdout, result)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runStatus(args []string) (status.Report, error) {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	agent := fs.String("agent", string(model.AgentCodex), "Agent to inspect")
	if err := fs.Parse(args); err != nil {
		return status.Report{}, err
	}
	agentID, err := parseKnownAgent(*agent)
	if err != nil {
		return status.Report{}, err
	}
	return status.BuildForAgent(Version, agentID)
}

func runDoctorValue(args []string) (any, error) {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	agent := fs.String("agent", string(model.AgentCodex), "Agent to inspect")
	fix := fs.Bool("fix", false, "Install missing components")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	agentID, err := parseKnownAgent(*agent)
	if err != nil {
		return nil, err
	}
	if *fix {
		if !model.IsAgentInstallable(agentID) {
			return nil, installableAgentError(*agent)
		}
		return doctor.FixForAgent(Version, agentID)
	}
	return doctor.RunForAgent(Version, agentID)
}

func runFlow(args []string) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("missing flow command; supported: status, quick")
	}
	switch args[0] {
	case "status":
		fs := flag.NewFlagSet("flow status", flag.ContinueOnError)
		if err := fs.Parse(args[1:]); err != nil {
			return nil, err
		}
		workDir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		return flowstate.Build(workDir)
	case "quick":
		if len(args) < 2 {
			return nil, fmt.Errorf("usage: nea-ai flow quick <change-name> [--title ...] [--objective ...]")
		}
		changeName := args[1]
		fs := flag.NewFlagSet("flow quick", flag.ContinueOnError)
		title := fs.String("title", "", "Quick title")
		objective := fs.String("objective", "", "Quick objective")
		files := fs.String("files", "", "Comma-separated affected files")
		verify := fs.String("verify", "", "Comma-separated verification commands")
		if err := fs.Parse(args[2:]); err != nil {
			return nil, err
		}
		workDir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		return flowstate.Quick(workDir, flowstate.QuickOptions{
			Name:         changeName,
			Title:        *title,
			Objective:    *objective,
			Files:        splitComponents(*files),
			Verification: splitComponents(*verify),
		})
	default:
		return nil, fmt.Errorf("unsupported flow command %q; supported: status, quick", args[0])
	}
}

type InstallReport struct {
	Components map[model.ComponentID]any `json:"components"`
}

func runInstall(args []string) (InstallReport, error) {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	agent := fs.String("agent", string(model.AgentCodex), "Agent to configure")
	componentList := fs.String("components", string(model.ComponentBrain), "Comma-separated components")
	if err := fs.Parse(args); err != nil {
		return InstallReport{}, err
	}
	agentID, err := parseInstallableAgent(*agent)
	if err != nil {
		return InstallReport{}, err
	}
	paths, err := system.ResolvePaths()
	if err != nil {
		return InstallReport{}, err
	}
	registry := components.DefaultRegistry()
	ctx := components.ContextFromPaths(paths, agentID)
	report := InstallReport{Components: map[model.ComponentID]any{}}
	for _, component := range splitComponents(*componentList) {
		id := model.ComponentID(component)
		installer, ok := registry.Get(id)
		if !ok {
			return InstallReport{}, fmt.Errorf("unsupported component %q", component)
		}
		result, err := installer.Install(ctx)
		if err != nil {
			return InstallReport{}, err
		}
		report.Components[id] = result
	}
	return report, nil
}

type UninstallReport struct {
	Components map[model.ComponentID]any `json:"components"`
}

func runUninstall(args []string) (UninstallReport, error) {
	fs := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	agent := fs.String("agent", string(model.AgentCodex), "Agent to configure")
	componentList := fs.String("components", string(model.ComponentBrain), "Comma-separated components")
	if err := fs.Parse(args); err != nil {
		return UninstallReport{}, err
	}
	agentID, err := parseInstallableAgent(*agent)
	if err != nil {
		return UninstallReport{}, err
	}
	paths, err := system.ResolvePaths()
	if err != nil {
		return UninstallReport{}, err
	}
	registry := components.DefaultRegistry()
	ctx := components.ContextFromPaths(paths, agentID)
	report := UninstallReport{Components: map[model.ComponentID]any{}}
	for _, component := range splitComponents(*componentList) {
		id := model.ComponentID(component)
		installer, ok := registry.Get(id)
		if !ok {
			return UninstallReport{}, fmt.Errorf("unsupported component %q", component)
		}
		result, err := installer.Uninstall(ctx)
		if err != nil {
			return UninstallReport{}, err
		}
		report.Components[id] = result
	}
	return report, nil
}

func parseKnownAgent(raw string) (model.AgentID, error) {
	id := model.AgentID(raw)
	if !model.IsAgentKnown(id) {
		return "", fmt.Errorf("unknown agent %q; known: %s", raw, joinAgents(model.SupportedAgents()))
	}
	return id, nil
}

func parseInstallableAgent(raw string) (model.AgentID, error) {
	id := model.AgentID(raw)
	if !model.IsAgentInstallable(id) {
		return "", installableAgentError(raw)
	}
	return id, nil
}

func installableAgentError(raw string) error {
	return fmt.Errorf("agent %q is not installable yet; supported: %s", raw, joinAgents(model.InstallableAgents()))
}

func joinAgents(ids []model.AgentID) string {
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		names = append(names, string(id))
	}
	return strings.Join(names, ", ")
}

func splitComponents(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func writeJSON(stdout io.Writer, value any) error {
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}

func printHelp(stdout io.Writer) {
	_, _ = fmt.Fprintln(stdout, `NEA AI

Usage:
  nea-ai version
  nea-ai status [--agent codex|claude-code|opencode|cursor|vscode|gemini-cli]
  nea-ai doctor [--fix] [--agent codex|claude-code|opencode]
  nea-ai init
  nea-ai flow status
  nea-ai flow quick <change-name> [--title "..."] [--objective "..."]
  nea-ai install --agent codex|claude-code|opencode --components brain,flow
  nea-ai uninstall --agent codex|claude-code|opencode --components brain,flow

All commands emit JSON on stdout. Foundation commands implemented: version,
status, doctor, init, install/uninstall (brain,flow) and flow (status, quick)
for codex, claude-code, and opencode.`)
}
