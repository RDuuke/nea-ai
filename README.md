# NEA AI

NEA AI es un control plane local-first para agentes de desarrollo con IA.

Su objetivo es preparar agentes existentes con memoria persistente, flujos de
trabajo estructurados, estado OpenSpec, configuracion MCP, verificacion y
recuperacion en un solo binario.

```text
NEA AI = installer + NeaBrain memory + Flow-NEA + OpenSpec + doctor
```

## Que Hace

- Configura NeaBrain como servidor MCP para agentes soportados.
- Instala skills, prompts y comandos de Flow-NEA.
- Inicializa estado OpenSpec en proyectos.
- Detecta agentes instalados y salud de componentes.
- Repara configuracion faltante con `doctor --fix`.
- Crea backups antes de tocar archivos de configuracion.
- Desinstala solo entradas administradas por NEA AI.

## Agentes Soportados

Soporte actual:

- Codex
- OpenCode
- Claude Code

Planeado:

- Cursor
- VS Code
- Gemini CLI

## Componentes

- `brain`: instala/configura NeaBrain como MCP server.
- `flow`: instala Flow-NEA skills, prompts y commands para el agente.

Repos base:

- NeaBrain: https://github.com/RDuuke/nea-brain
- Flow-NEA: https://github.com/RDuuke/sdd-nea-flow

## Instalacion

Desde este repositorio:

```bash
go run ./cmd/nea-ai install --agent codex --components brain,flow
go run ./cmd/nea-ai install --agent opencode --components brain,flow
go run ./cmd/nea-ai install --agent claude-code --components brain,flow
```

## Inicializar Un Proyecto

```bash
go run ./cmd/nea-ai init
```

Crea:

```text
openspec/
  config.yaml
  changes/
    .status.yaml
```

## Estado

Todos los comandos emiten JSON en stdout.

```bash
go run ./cmd/nea-ai status --agent codex
go run ./cmd/nea-ai status --agent opencode
go run ./cmd/nea-ai status --agent claude-code
```

## Flow

Ver estado OpenSpec/Flow-NEA del proyecto actual:

```bash
go run ./cmd/nea-ai flow status
```

Crear un quick blueprint para un cambio chico:

```bash
go run ./cmd/nea-ai flow quick fix-readme --title "ajustar readme" --objective "Mejorar documentacion publica"
```

Crear artefactos OpenSpec para fases de Flow-NEA:

```bash
go run ./cmd/nea-ai flow explore add-feature --objective "Understand the change"
go run ./cmd/nea-ai flow propose add-feature --summary "Bounded implementation plan"
go run ./cmd/nea-ai flow continue
go run ./cmd/nea-ai flow verify add-feature --summary "CI passed" --commands "go test ./..."
```

## Doctor

Validar instalacion:

```bash
go run ./cmd/nea-ai doctor --agent opencode
```

Reparar componentes faltantes:

```bash
go run ./cmd/nea-ai doctor --fix --agent opencode
```

## Desinstalar

```bash
go run ./cmd/nea-ai uninstall --agent opencode --components brain,flow
```

`uninstall` elimina entradas y archivos conocidos administrados por NEA AI. No
borra configuracion ajena del usuario.

## Build De Release

Build local con version inyectada:

```bash
go build -ldflags "-X nea-ai/internal/app.Version=v0.2.0" -o dist/nea-ai-windows-amd64.exe ./cmd/nea-ai
```

O via Makefile:

```bash
make build VERSION=v0.2.0
```

### Release Automatico

Al empujar un tag `v*` el workflow `.github/workflows/release.yml` corre
GoReleaser y publica artefactos para `linux|darwin|windows` x `amd64|arm64`,
mas `checksums.txt`.

```bash
git tag v0.2.0
git push origin v0.2.0
```

Validar la configuracion antes de tagear:

```bash
goreleaser check
goreleaser release --snapshot --clean
```

## Alcance Actual

NEA AI no es otro agente de codigo. Es el runtime/control plane que prepara los
agentes existentes con memoria, flujo, verificacion y recuperacion.

El MVP actual cubre:

- instalacion multi-agente
- configuracion MCP NeaBrain
- instalacion Flow-NEA
- bootstrap OpenSpec
- `flow status`
- `flow quick`
- `flow explore`
- `flow propose`
- `flow continue`
- `flow verify`
- `status`
- `doctor`
- `doctor --fix`
- `uninstall`

## Roadmap Corto

Siguientes bloques:

- `flow` commands desde `nea-ai`
- ping real de MCP/NeaBrain
- soporte Cursor, VS Code y Gemini CLI
- releases multiplataforma
- TUI/dashboard local
