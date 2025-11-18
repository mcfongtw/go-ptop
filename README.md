# go-ptop (POC/MVP)

go-ptop is an experimental Go tool for inspecting JVM memory/thread usage by correlating `/proc/<pid>/smaps`, kernel thread stats, and JVM thread dumps. We are currently in a proof-of-concept (POC) phase focused on Linux-only support with a single CLI table output.

## Current Status
- **Scope**: MVP features M1–M7 defined in [doc/prd/requirement.md](doc/prd/requirement.md#5-functional-requirements).
- **Platforms**: Linux only. macOS/Windows support will be evaluated after MVP validation.
- **Outputs**: Basic table printed to stdout; JSON/TUI modes planned for later phases.

## Development Notes
1. Read [doc/prd/prompt.md](doc/prd/prompt.md) before making changes. It contains workflow and testing conventions.
2. High-level architecture concepts are documented in [doc/design.md](doc/design.md) but only partially implemented during MVP.
3. Requirement/Design/BDD docs must stay consistent once the project leaves POC; for MVP we focus on verifying the concept.

## Running
```bash
go build -o go-ptop
./go-ptop <pid>
```

The target JVM must run on Linux with attach permissions (same user). The tool emits minimal error messages if attach fails or `/proc` access is denied.
