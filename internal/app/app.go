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
		report, err := status.Build(Version)
		if err != nil {
			return err
		}
		return writeJSON(stdout, report)
	case "doctor":
		report, err := doctor.Run(Version)
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
	default:
		return fmt.Errorf("unknown command %q", args[0])
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
  nea-ai status --json
  nea-ai doctor
  nea-ai init
  nea-ai install --agent codex --components brain,flow

Foundation commands are implemented: version, status, doctor, init, install brain for codex.`)
}
