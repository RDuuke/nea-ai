# NEA AI

Control plane local-first para agentes de desarrollo.

MVP actual:

```bash
go run ./cmd/nea-ai version
go run ./cmd/nea-ai status --json
go run ./cmd/nea-ai doctor
go run ./cmd/nea-ai init
go run ./cmd/nea-ai install --agent codex --components brain,flow
```

Objetivo:

```text
NEA AI = installer + NeaBrain memory + NEA Flow + OpenSpec + doctor
```

Ver plan: `NEA_AI_PROJECT_PLAN.md`.
