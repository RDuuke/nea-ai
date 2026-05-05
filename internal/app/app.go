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
		report, err := runDoctor(args[1:])
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
	_ = fs.Bool("json", true, "Write JSON output")
	if err := fs.Parse(args); err != nil {
		return status.Report{}, err
	}
	return status.BuildForAgent(Version, model.AgentID(*agent))
}

func runDoctor(args []string) (doctor.Report, error) {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	agent := fs.String("agent", string(model.AgentCodex), "Agent to inspect")
	if err := fs.Parse(args); err != nil {
		return doctor.Report{}, err
	}
	return doctor.RunForAgent(Version, model.AgentID(*agent))
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
	paths, err := system.ResolvePaths()
	if err != nil {
		return InstallReport{}, err
	}
	registry := components.DefaultRegistry()
	ctx := components.ContextFromPaths(paths, model.AgentID(*agent))
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
	paths, err := system.ResolvePaths()
	if err != nil {
		return UninstallReport{}, err
	}
	registry := components.DefaultRegistry()
	ctx := components.ContextFromPaths(paths, model.AgentID(*agent))
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
  nea-ai status --json [--agent codex|opencode|claude-code]
  nea-ai doctor [--agent codex|opencode|claude-code]
  nea-ai init
  nea-ai install --agent codex|opencode|claude-code --components brain,flow
  nea-ai uninstall --agent codex|opencode|claude-code --components brain,flow

Foundation commands are implemented: version, status, doctor, init, install/uninstall brain/flow for codex, opencode, and claude-code.`)
}
