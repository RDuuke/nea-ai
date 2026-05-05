# NEA AI

NEA AI is a local-first control plane for AI development agents.

It installs and repairs the pieces needed to use persistent memory, structured
Flow-NEA workflows, OpenSpec project state, and MCP configuration across
supported coding agents.

## What It Does

- Configures NeaBrain MCP for supported agents.
- Installs Flow-NEA skills, prompts, and commands.
- Initializes OpenSpec project state.
- Detects installed agents and component health.
- Repairs missing agent configuration with `doctor --fix`.
- Backs up touched config files before writing.

## Supported Agents

Current support:

- Codex
- OpenCode
- Claude Code

Planned:

- Cursor
- VS Code
- Gemini CLI

## Install Components

From this repository:

```bash
go run ./cmd/nea-ai install --agent codex --components brain,flow
go run ./cmd/nea-ai install --agent opencode --components brain,flow
go run ./cmd/nea-ai install --agent claude-code --components brain,flow
```

Installed components:

- `brain`: configures NeaBrain as an MCP server.
- `flow`: installs Flow-NEA skills/prompts/commands for the selected agent.

## Project Init

Initialize OpenSpec state in the current project:

```bash
go run ./cmd/nea-ai init
```

This creates:

```text
openspec/
  config.yaml
  changes/
    .status.yaml
```

## Status

Inspect machine and project state:

```bash
go run ./cmd/nea-ai status --agent codex --json
go run ./cmd/nea-ai status --agent opencode --json
go run ./cmd/nea-ai status --agent claude-code --json
```

## Doctor

Validate an agent setup:

```bash
go run ./cmd/nea-ai doctor --agent opencode
```

Repair missing components:

```bash
go run ./cmd/nea-ai doctor --fix --agent opencode
```

## Uninstall

Remove NEA AI managed configuration for an agent:

```bash
go run ./cmd/nea-ai uninstall --agent opencode --components brain,flow
```

Uninstall removes NEA-managed entries and known Flow-NEA files. It does not
remove unrelated user configuration.

## Release Build

Windows example:

```bash
go build -ldflags "-X nea-ai/internal/app.Version=v0.2.0" -o dist/nea-ai-windows-amd64.exe ./cmd/nea-ai
```

## Current Scope

NEA AI is not a new coding agent. It is the runtime/control plane that prepares
existing agents with memory, workflow, verification, and recovery support.

Current MVP:

```text
NEA AI = installer + NeaBrain memory + Flow-NEA + OpenSpec + doctor
```
