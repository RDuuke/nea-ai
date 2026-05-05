# Proyecto Derivado: Gentle AI Control Center

## Base Revisada

Repositorio: `Gentleman-Programming/gentle-ai`

Producto actual: CLI/TUI en Go que configura ecosistemas de agentes AI. Instala o sincroniza agentes, skills, SDD, Engram, MCP, persona, permisos, backups, actualizaciones y perfiles de modelo.

Stack:
- Go `1.24.2`
- Bubble Tea / Bubbles / Lipgloss para TUI
- GoReleaser para releases
- Tests Go + golden fixtures + E2E Docker

## Lectura De Arquitectura

Estructura principal:
- `cmd/gentle-ai`: entrypoint
- `internal/app`: dispatch CLI/TUI, self-update, comandos
- `internal/cli`: flags, install/sync/uninstall/restore, pipeline real
- `internal/tui`: pantallas Bubble Tea
- `internal/agents`: adapters por agente
- `internal/components`: instaladores/injectores de Engram, SDD, skills, MCP, persona, permisos, theme, GGA
- `internal/planner`: grafo de dependencias y orden de componentes
- `internal/pipeline`: ejecución por etapas + rollback
- `internal/backup`: snapshots, restore, retention
- `internal/update`: update/upgrade
- `internal/assets`: templates embebidos
- `openspec`: cambios y specs existentes

## Oportunidad De Proyecto

Crear una capa de gestión visual encima de Gentle AI: **Gentle AI Control Center**.

Objetivo: convertir el instalador/configurador en un centro de control para ver, editar, validar y sincronizar configuración de agentes AI en una máquina o equipo.

No reemplaza CLI actual. La envuelve.

## Problema

Gentle AI ya hace mucho, pero usuario necesita saber:
- qué agentes están instalados
- qué componentes activos tiene cada agente
- qué archivos fueron tocados
- qué modelo usa cada fase SDD
- qué backups existen
- qué cambió entre syncs
- qué está roto y cómo arreglarlo

Hoy eso vive repartido entre TUI, CLI, docs, state y filesystem.

## Propuesta

Agregar módulo `control-center` con vista web local o TUI avanzada:
- dashboard de agentes instalados
- matriz agente x componente
- editor de perfiles SDD/modelos
- explorador de backups y restore
- health checks ejecutables
- preview de `sync` antes de escribir
- export/import de configuración de equipo

## MVP

### 1. Inventory

Comando:

```bash
gentle-ai status --json
```

Debe devolver:
- agentes detectados
- componentes detectados
- rutas de config por agente
- estado de Engram/GGA
- versión de gentle-ai
- último backup

### 2. Dry-Run Diff

Comando:

```bash
gentle-ai sync --dry-run --diff
```

Debe mostrar:
- archivos que cambiarían
- secciones que se insertarían
- backups que se crearían
- warnings de colisión

### 3. Profile Manager

Comandos:

```bash
gentle-ai profiles list
gentle-ai profiles create <name>
gentle-ai profiles set <name> <phase> <provider/model>
gentle-ai profiles delete <name>
```

Base: lógica actual de OpenCode profiles en `internal/model`, `internal/components/sdd`, `internal/opencode`.

### 4. Health Repair

Comando:

```bash
gentle-ai doctor
gentle-ai doctor --fix
```

Debe revisar:
- binaries en PATH
- archivos requeridos
- MCP JSON/TOML válido
- permisos/config seguros
- backups disponibles
- state.json consistente

## Roadmap

Fase 1: estado legible
- implementar `status --json`
- extraer detectores reutilizables por componente
- añadir tests por agente
- documentar schema JSON

Fase 2: diff seguro
- crear writer en modo preview
- centralizar file merge diff
- snapshot tests con golden files
- mostrar resumen en CLI/TUI

Fase 3: perfiles
- normalizar perfiles SDD cross-agent, no solo OpenCode
- persistir perfiles en state/config
- migración de perfiles existentes
- comandos CRUD

Fase 4: doctor
- reutilizar `verify`
- separar checks hard/soft
- añadir fixers idempotentes
- reportes accionables

Fase 5: UI
- TUI dashboard dentro de Bubble Tea
- opcional: server local `gentle-ai ui`
- export/import team profile

## Backlog Inicial

Prioridad alta:
- `internal/status`: nuevo paquete con `BuildReport(homeDir, detection)`
- `gentle-ai status --json`
- tests para status con temp HOME
- schema documentado en `docs/status.md`
- `doctor` basado en `internal/verify`

Prioridad media:
- diff preview para `filemerge`
- profile CRUD CLI
- TUI screen para status
- backup browser mejorado con filtros

Prioridad baja:
- web local
- team profile registry
- remote sync de configuración
- marketplace de skills

## Riesgos

- Mucha lógica actual escribe directo a filesystem. Para diff real se necesita separar render plan de apply.
- Adapters tienen estrategias distintas. Status común necesita contrato estable.
- Windows PATH y permisos pueden dar falsos negativos.
- Profile manager cross-agent puede chocar con capacidades nativas diferentes.

## Primer Corte Recomendado

Construir `status --json` primero.

Razón: bajo riesgo, alto valor, desbloquea dashboard, doctor, diff, soporte y documentación.

Definition of Done:
- comando `gentle-ai status --json`
- cubre 12 agentes definidos en `internal/model/types.go`
- no escribe archivos
- salida estable con tests golden
- docs con ejemplo real

