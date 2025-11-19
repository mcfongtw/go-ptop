# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

go-ptop is a Java process analyzer that provides terminal-based monitoring of Java thread and memory usage. It connects to Java processes via JVM Attach API to display real-time thread and memory information, correlating data from multiple sources: /proc filesystem (smaps), JVM Native Memory Tracking (NMT), and kernel thread information.

## Architecture

The codebase follows an interface-driven, package-based architecture designed for testability and cross-platform extensibility.

### Package Structure

```
go-ptop/
├── main.go                     # Entry point, CLI arg parsing, error handling
├── pkg/
│   ├── memory/                 # Memory reader (smaps, pagemap, NUMA)
│   │   ├── type.go             # MemoryReader interface, ProcessMemorySegment
│   │   ├── memory_linux.go     # Linux implementation
│   │   ├── memory_stub.go      # Stub for unsupported platforms
│   │   ├── smaps_parser.go     # /proc/pid/smaps parser
│   │   └── smaps_parser_test.go
│   ├── jvm/                    # JVM client (attach API, NMT, thread dumps)
│   │   ├── type.go             # JVMClient interface, JavaThread, NMTReport
│   │   ├── client_linux.go     # Linux implementation (Unix sockets)
│   │   ├── client_stub.go      # Stub for unsupported platforms
│   │   ├── thread_parser.go    # Thread dump parser
│   │   └── thread_parser_test.go
│   ├── proc/                   # Process info reader (kernel threads, I/O stats)
│   │   ├── type.go             # ProcReader interface, KernelThread, IOStats
│   │   ├── reader_linux.go     # Linux implementation
│   │   ├── reader_stub.go      # Stub for unsupported platforms
│   │   ├── stat_parser.go      # /proc/pid/task/tid/stat parser
│   │   └── stat_parser_test.go
│   ├── analyzer/               # Correlator for data sources
│   │   ├── type.go             # Analyzer interface, AnalysisResult
│   │   ├── simple_analyzer.go  # Basic implementation
│   │   └── simple_analyzer_test.go
│   ├── formatter/              # Output formatters
│   │   ├── type.go             # Formatter interface, TableFormatter
│   │   └── formatter.go        # Table formatter implementation
│   ├── config/                 # Configuration management
│   │   ├── type.go             # Config struct
│   │   └── config.go
│   ├── ui/                     # TUI interface (optional)
│   │   ├── type.go
│   │   └── ui.go
│   └── util/                   # String formatting utilities
│       └── format.go
├── doc/
│   ├── design.md               # Full architecture design document
│   ├── development.md          # Development notes
│   └── prd/                    # Product requirements
└── go.mod
```

### Key Interfaces

**MemoryReader** (`pkg/memory/type.go`):
- `ReadSMaps(pid int32) ([]ProcessMemorySegment, error)` - Parse /proc/pid/smaps
- `ReadPageMap(pid int32, virtAddr, length uint64) ([]PageInfo, error)` - Page-level info (root)
- `ReadNUMAMaps(pid int32) ([]NUMASegment, error)` - NUMA allocations (root)

**JVMClient** (`pkg/jvm/type.go`):
- `Connect(pid int32) error` - Establish JVM connection
- `GetThreadDump() (*ThreadDump, error)` - Full thread dump
- `GetNMT(detail bool) (*NMTReport, error)` - Native Memory Tracking report
- `GetHeapInfo() (*HeapInfo, error)` - Heap metrics
- `ExecuteCommand(command string, args ...string) (string, error)` - Arbitrary jcmd

**ProcReader** (`pkg/proc/type.go`):
- `Threads(pid int32) ([]KernelThread, error)` - List kernel threads
- `ThreadStat(pid int32, tid int32) (*ThreadStat, error)` - Thread stat info
- `ThreadIO(pid int32, tid int32) (*IOStats, error)` - Per-thread I/O stats

**Analyzer** (`pkg/analyzer/type.go`):
- `Analyze(pid int32) (*AnalysisResult, error)` - Full analysis
- `CorrelateThreads(...)` - Match Java threads to kernel threads and memory
- `ClassifyMemory(...)` - Determine memory segment purposes

**Formatter** (`pkg/formatter/type.go`):
- `Format(result *AnalysisResult) (string, error)` - Render output

### Key Data Structures

- `ProcessMemorySegment`: Memory mapping from /proc/pid/smaps with RSS, PSS, permissions
- `JavaThread`: Java thread metadata (name, OS ID, state, stack trace)
- `KernelThread`: Kernel thread info (TID, state, start stack)
- `IOStats`: Per-thread I/O counters (read/write counts and bytes)
- `AnalysisResult`: Complete analysis with categorized memory and thread correlations
- `NMTReport`: Parsed Native Memory Tracking data

### Core Workflow

1. Parse target Java process PID from command line
2. Create readers: `memory.NewMemoryReader()`, `jvm.NewClient()`, `proc.NewReader()`
3. Create analyzer with `analyzer.New(memReader, jvmClient, procReader)`
4. Call `analyzer.Analyze(pid)` which:
   - Reads process memory mappings from /proc/pid/smaps
   - Connects to JVM and retrieves thread dump and NMT data
   - Reads kernel thread information from /proc/pid/task
   - Correlates Java threads with kernel threads and memory segments
   - Classifies memory by purpose (heap, metaspace, thread stacks, etc.)
5. Format output with `formatter.NewTableFormatter().Format(result)`

### Error Handling

The codebase uses sentinel errors for platform support:
- `memory.ErrUnsupported` - Memory reader not available
- `jvm.ErrUnsupported` - JVM attach not available
- `proc.ErrUnsupported` - Proc reader not available

Custom error types in `pkg/analyzer/type.go`:
- `ErrJVMNotFound` - Target process not found
- `ErrNMTNotEnabled` - NMT not enabled on target JVM
- `ErrRequiresRoot` - Feature needs root privileges
- `ErrAttachFailed` - JVM attach failed

## Development Commands

### Build and Run
```bash
# Download dependencies
go mod tidy

# Build application
go build -o go-ptop

# Run application (requires Java process PID)
./go-ptop <pid>
```

### Testing
```bash
# Run all tests (requires unit build tag)
go test ./... -tags=unit

# Run tests with verbose output
go test -v ./... -tags=unit

# Run tests for a specific package
go test -v ./pkg/memory/... -tags=unit
go test -v ./pkg/jvm/... -tags=unit
go test -v ./pkg/proc/... -tags=unit
go test -v ./pkg/analyzer/... -tags=unit
```

Note: Test files use `//go:build unit` build tag to separate unit tests from integration tests.

### Dependencies

Direct dependencies (from go.mod):
- `github.com/shirou/gopsutil` - Process utilities
- `golang.org/x/sys` - Unix system calls

## Platform Requirements

**Linux (Primary - Full Support)**:
- All features available
- Requires access to /proc filesystem
- JVM attach via Unix domain sockets at `/tmp/.java_pid<pid>`
- Root required for pagemap and NUMA features

**macOS (Planned - Good Support)**:
- JVM attach works (same Unix socket protocol)
- Memory via Mach VM APIs instead of /proc
- No PSS, pagemap, or NUMA support

**Windows (Planned - Basic Support)**:
- JVM attach via named pipes
- Memory via Win32 APIs
- No per-thread I/O stats

## Design Principles

### Separation of Concerns
Each package has a single responsibility:
- `memory` - Read /proc memory info
- `jvm` - Communicate with JVM
- `proc` - Read /proc process info
- `analyzer` - Correlate and classify data
- `formatter` - Format output

### Interface-Driven Design
All major components are interfaces enabling:
- Easy testing with mocks
- Multiple implementations (platform-specific)
- Clear contracts between components

### Graceful Degradation
- NMT not enabled -> Fall back to smaps-only analysis
- No root access -> Skip pagemap/numa features
- Unsupported platform -> Return sentinel error

### Platform Abstraction
Use build tags for platform-specific code:
- `*_linux.go` - Linux implementation
- `*_stub.go` - Stub for unsupported platforms (returns ErrUnsupported)

## Testing Conventions

- Test files use `*_test.go` suffix
- Test setup helpers in `test_setup.go` files
- Use table-driven tests where appropriate
- Mock interfaces for unit testing

## Code Conventions

- Use Go standard formatting (`go fmt`)
- Error messages should be lowercase without trailing punctuation
- Prefer returning errors over panicking
- Use sentinel errors for expected failure conditions
- Document exported types and functions

## Common Tasks

### Adding a New Platform

1. Create `*_<platform>.go` files in each package with build tag
2. Implement the interface (MemoryReader, JVMClient, ProcReader)
3. Update `New*()` factory functions to return platform-specific implementation

### Adding a New Output Format

1. Create new formatter type in `pkg/formatter/`
2. Implement the `Formatter` interface
3. Add factory function (e.g., `NewJSONFormatter()`)

### Adding New Analysis Features

1. Extend `AnalysisResult` struct in `pkg/analyzer/type.go`
2. Update `simple_analyzer.go` to populate new fields
3. Update formatters to display new data

## Reference Documentation

See `doc/design.md` for the complete architecture proposal including:
- Detailed interface definitions
- Migration path from old architecture
- Cross-platform feasibility analysis
- Future enhancement plans
