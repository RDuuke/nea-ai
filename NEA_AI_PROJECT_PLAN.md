# NEA AI - Proyecto Propio

## Repos Revisados

- `C:\Users\juandg\Desktop\RDuuqe\nea-brain`
- `C:\Users\juandg\Desktop\RDuuqe\tdd-nea-flow`

## Lectura Corta

`nea-brain` ya es el cerebro: memoria local-first para agentes IA.

`tdd-nea-flow` ya es el metodo: flujo SDD con orquestador, sub-agentes, fases, OpenSpec y gates.

NEA AI propio debe unir esas 2 piezas con un instalador/control-plane:

```text
NEA AI = installer + flow + memory + agent control plane
```

## Que Ya Existe

### NeaBrain

Stack:
- Go
- SQLite + FTS5
- MCP stdio
- CLI
- HTTP
- TUI Bubble Tea
- sync por chunks JSONL.gz

Arquitectura:
- `domain`: entidades y reglas
- `ports`: interfaces inbound/outbound
- `adapters`: CLI, HTTP, MCP, TUI, SQLite
- `app`: bootstrap
- `sync`: export/import local-first

Valor:
- memoria persistente
- busqueda entre sesiones
- captura pasiva
- resumen de sesion
- contexto por proyecto/topic
- MCP tools para agentes

### TDD NEA Flow

Stack:
- Markdown skills
- ejemplos por herramienta
- scripts de instalacion
- OpenSpec como artifact store

Valor:
- flujo por fases
- orquestador liviano
- sub-agentes por fase
- artifacts versionables
- retries, approvals, verify, archive
- quick path para fixes chicos

## Producto Propuesto

Nombre: **NEA AI**

Descripcion:
CLI/TUI que instala y opera un sistema AI local-first para desarrollo: memoria persistente, flujo SDD, skills, configuracion de agentes, backups, health checks y dashboard.

No debe ser "otro agente". Debe ser el runtime/control-plane que mejora agentes existentes.

## Arquitectura Target

```text
cmd/nea-ai
internal/
  app/                 dispatch CLI/TUI
  model/               agentes, componentes, perfiles, fases
  catalog/             registry de agentes, skills, componentes
  system/              OS, paths, deps
  installer/           install/sync/uninstall
  pipeline/            stages + rollback
  backup/              snapshots config
  agents/              adapters claude/opencode/codex/cursor/vscode/gemini
  components/
    brain/             instala/configura NeaBrain
    flow/              instala skills tdd-nea-flow
    mcp/               config MCP por agente
    persona/           instrucciones NEA
    permissions/       reglas seguras
  flow/                estado OpenSpec, phase runner, status
  memory/              wrapper/cliente NeaBrain
  doctor/              health checks y fixers
  tui/                 dashboard
assets/
  skills/              copia vendorizada de tdd-nea-flow/skills
  prompts/             AGENTS/CLAUDE/GEMINI/opencode examples
```

## Componentes MVP

### 1. `nea-ai init`

Inicializa proyecto:
- crea `openspec/`
- escribe `openspec/config.yaml`
- activa `experimental.neabrain: true`
- detecta stack
- registra proyecto en NeaBrain

### 2. `nea-ai install`

Instala en maquina:
- binario `neabrain`
- skills de `tdd-nea-flow`
- MCP config para agente seleccionado
- prompts base NEA
- backup antes de tocar configs

Agentes iniciales:
- Codex
- Claude Code
- OpenCode
- Cursor

### 3. `nea-ai status`

Muestra:
- agentes detectados
- NeaBrain disponible
- MCP conectado
- skills instaladas
- OpenSpec presente
- cambios activos
- ultima memoria capturada
- backups

### 4. `nea-ai flow`

Controla fases:
- `nea-ai flow init`
- `nea-ai flow explore <name>`
- `nea-ai flow quick <name>`
- `nea-ai flow continue`
- `nea-ai flow verify`

Al inicio no necesita ejecutar sub-agentes internamente. Puede instalar prompts/skills y dejar que el agente los use.

### 5. `nea-ai doctor`

Valida:
- DB NeaBrain abre
- SQLite migrations ok
- MCP responde `tools/list`
- skills existen
- OpenSpec status valido
- config JSON/TOML valida

## Decision Tecnica

Base recomendada: Go CLI, no Node.

Razon:
- `nea-brain` ya es Go
- single binary simplifica Windows
- Go facilita CLI/TUI futura con un binario unico

## Reuso Recomendado

Copiar/adaptar de `nea-brain`:
- MCP server patterns
- setup por agente
- sync chunks
- config precedence
- hexagonal style

Copiar/adaptar de `tdd-nea-flow`:
- skills completas
- examples por herramienta
- persistence contract
- OpenSpec structure
- phase output contract

## Lo Que No Conviene Hacer

- No mezclar todo dentro de `nea-brain`. Mantener memoria como componente independiente.
- No convertir `tdd-nea-flow` en app. Mantenerlo como paquete de skills/prompts.
- No crear agente AI propio desde cero. Mejor configurar agentes existentes.
- No meter web UI en MVP. Primero CLI/TUI.

## Roadmap

### Fase 1 - Foundation

- crear repo `nea-ai`
- Go module
- CLI skeleton
- detector OS/home/config paths
- command `version`, `help`, `status`

### Fase 2 - Installer

- agent adapters para Codex/Claude/OpenCode/Cursor
- backup antes de escribir
- install/uninstall de MCP NeaBrain
- install de flow skills

### Fase 3 - Project Init

- `openspec/` bootstrap
- `.status.yaml`
- config `experimental.neabrain`
- captura memoria inicial en NeaBrain

### Fase 4 - Doctor

- health checks
- MCP ping
- status JSON
- repair simple

### Fase 5 - TUI

- dashboard local
- agentes
- componentes
- memoria
- cambios OpenSpec

## MVP Realista

Primer entregable:

```bash
nea-ai install --agent codex --components brain,flow
nea-ai init
nea-ai status --json
nea-ai doctor
```

Definition of Done:
- instala sin romper configs
- crea backup
- configura NeaBrain MCP
- copia skills flow
- inicializa OpenSpec
- `status` reporta todo
- `doctor` detecta fallos claros

## Proyecto En Una Frase

NEA AI debe ser el **control plane local-first para agentes de desarrollo**: memoria, flujo, configuracion, verificacion y recuperacion en un solo binario.
