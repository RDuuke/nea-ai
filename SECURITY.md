# Security Policy

## Supported Versions

NEA AI is pre-1.0. Only the latest tagged release receives security fixes.

## Reporting a Vulnerability

Please **do not** open a public GitHub issue for security problems.

Email the maintainer at `juandg@iris.com.co` with:

- A description of the issue and its impact.
- Steps to reproduce, ideally with a minimal proof-of-concept.
- The version (`nea-ai version`) and OS where it was observed.

You will receive an acknowledgement within 5 business days. Confirmed issues
will be triaged, fixed in a private branch, and disclosed in the release notes
once a patched version is published.

## Scope

In scope:

- The `nea-ai` binary and its CLI surface.
- `internal/` packages of this repository.

Out of scope:

- NeaBrain (`https://github.com/RDuuke/nea-brain`) — report there.
- Flow-NEA skills (`https://github.com/RDuuke/sdd-nea-flow`) — report there.
- Third-party agents (Codex, OpenCode, Claude Code, Cursor, etc.).
