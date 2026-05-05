# Contributing

NEA AI is a local-first control plane for AI development agents. Contributions
welcome via PRs against `main`.

## Requirements

- Go 1.24+
- `golangci-lint` v1.64.5 (or run via `make install-tools`)
- A POSIX shell or PowerShell

## Local Workflow

```bash
make tidy        # ensure go.mod/go.sum are clean
make vet         # go vet ./...
make test        # go test ./... -count=1
make lint        # golangci-lint
make build       # produces dist/nea-ai
```

To run with the local build:

```bash
make run ARGS="status --agent codex"
```

## Coding Conventions

- Format with `gofmt` (CI enforces this via `gofmt`/`goimports`).
- Imports: standard library, then third-party, then `nea-ai/...`.
- Errors: wrap with `fmt.Errorf("context: %w", err)` so callers can `errors.Is`.
- File operations: always use `filepath.Join` for cross-OS paths.
- New CLI flags: add validation in `internal/app/app.go` before delegating to
  the implementing package.

## Pull Requests

- One logical change per PR. Conventional Commits style for the title:
  `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`.
- Add tests for new logic. CI requires `go test ./...` and `golangci-lint` to
  pass on Linux, macOS, and Windows.
- Reference an issue if one exists.

## Reporting Bugs

Open a GitHub issue with: OS, Go version, command run, expected output,
actual output, and JSON of `nea-ai status --agent <agent>` if relevant.
