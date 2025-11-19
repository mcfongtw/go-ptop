# go-ptop Feature Requirements

## References
- [Prompt & Coding Instructions](./prompt.md)
- [Design Doc](../design.md)
- [BDD Feature Definition](../bdd/feature.md) <!-- Placeholder link per documentation practice -->

## 1. Problem Statement
Java engineers need a lightweight terminal tool to inspect JVM thread activity and memory footprint on live Linux systems. We are currently in a proof-of-concept (POC) phase where the objective is to validate the usefulness of correlating `/proc` smaps, kernel thread stats, and JVM thread dumps. The legacy prototype mixes UI, process parsing, and JVM attach logic in single files; for the MVP we only need the smallest slice that demonstrates the idea before investing in the full modular redesign.

## 2. Goals & Success Metrics
### MVP/POC Goals
1. **Validate attach + smaps correlation**: prove we can read `/proc/<pid>/smaps`, fetch a JVM thread dump via Attach API, and correlate thread stacks to memory segments on Linux.
2. **Deliver a minimal CLI table**: render correlated thread + memory data in a single table output so engineers can inspect top threads quickly.
3. **Provide basic error messaging**: missing PID, permission errors, or attach failures should surface in the CLI output.

Success criteria for MVP:
- Running `go-ptop <pid>` on Linux produces a table showing thread names, IDs, RSS, and IO counters.
- JVM attach and smaps parsing work end-to-end without panics.
- Inaccessible features (e.g., NMT, pagemap) clearly state that they are “future work / unsupported”.

### Post-MVP Aspirations
Once MVP proves useful, we will expand toward the full architecture described in [Design Doc](../design.md): multi-format outputs, classification, capability matrix, multi-platform support, structured logging, and documentation workflow.

## 3. Out of Scope
- Full Windows/macOS feature parity beyond the tiers defined in [Design Doc](../design.md#cross-platform-feasibility-analysis).
- Advanced diagnostics such as eBPF sampling, perf counters, or historical trend storage.
- Automated remediation or JVM tuning recommendations (observation only).

## 4. Personas & Use Cases
| Persona | Scenario | Requirement IDs |
| --- | --- | --- |
| JVM Operator | Diagnose production memory spike by correlating smaps regions with Java threads | F1, F3, F4, F5 |
| Performance Engineer | Capture JSON snapshot for offline analysis pipelines | F1, F6 |
| Support Engineer | Verify if JVM NMT is enabled and surface actionable error when not | F2, F3, F8 |

## 5. Functional Requirements
### MVP Scope (Phase 0 / POC)
| ID | Description | Acceptance Criteria | Dependencies |
| --- | --- | --- | --- |
| **M1. Process Targeting** | CLI accepts a single PID argument, validates it, and prints usage if missing. | `go-ptop <pid>` attaches to that process; invalid/missing PID results in a clear error. | main.go |
| **M2. SMaps Parsing (Linux)** | Minimal `memory` component reads `/proc/<pid>/smaps` and returns segment stats (start/end, perms, RSS, PSS). | Running against a Java process yields non-empty segment slice; errors bubble up if /proc access fails. | Linux-only |
| **M3. JVM Thread Dump (Linux Attach)** | Minimal `jvm` component triggers JVM thread dump via Attach API (SIGQUIT + socket) and extracts Java thread name + native TID mapping. | Correlated data includes Java thread names and OS TIDs; if attach fails, user sees descriptive message. | Linux JVM |
| **M4. Kernel Thread Stats** | Minimal `proc` component lists `/proc/<pid>/task/<tid>` entries and IO stats. | Output table shows TID, read/write counts (or zero if not available). | Linux-only |
| **M5. Basic Correlation** | Lightweight analyzer correlates Java threads to kernel threads and memory segments by stack addresses, returning RSS totals and per-thread segment info. | CLI table shows thread name, TID, RSS, IO counts; segments default to “unknown” classification. | Depends on M2–M4 |
| **M6. CLI Table Output** | Provide a simple table formatter (no TUI) which prints rows sorted by RSS or TID. | Running the tool prints a header + rows to stdout; no JSON/text yet. | Go standard libs |
| **M7. Error Messaging** | Minimal error handling that reports attach failures, permissions, or unsupported platform. | For common failure modes, tool exits with non-zero status and human-readable message; no structured logging yet. | All MVP components |

### Post-MVP (Future Phase) – documented for traceability
| ID | Description | Notes |
| --- | --- | --- |
| **F2**–**F8** | Same as previously listed (multi-format output, NMT, classification, capability matrix, structured logging) | kept for future but not part of MVP delivery |

## 6. Non-Functional Requirements
1. **Performance (MVP)**: Attach + smaps parse must complete within 3 seconds for processes with ≤300 smaps entries on Linux. Optimization and caching are deferred until post-MVP.
2. **Resource Usage**: Tool is read-only except for triggering JVM thread dump; no root features or pagemap access in MVP.
3. **Extensibility (Post-MVP)**: Interfaces are kept minimal but aligned with the future package structure to ease expansion.
4. **Documentation & Testing**: Minimal documentation for MVP (README + this requirements doc). Full doc/test practices from [Prompt](./prompt.md) kick in during formal phase.

## 7. Open Questions / Follow-ups
1. Confirm backward compatibility expectations for CLI flags once we graduate from MVP (Prompt §2.5). For MVP, we only support `go-ptop <pid>`.
2. Determine when structured logging (`common/logging`) investment begins; MVP uses simple stderr output.
3. Populate [BDD Feature Definition](../bdd/feature.md) after MVP validation to capture adoption criteria.

---
*This requirement document is to be kept in sync with [Doc/Design.md](../design.md) and the future BDD feature file per the documentation practice in [Prompt](./prompt.md).* 
