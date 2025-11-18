# go-ptop Architecture Design Proposal

## Overview

This document outlines the architectural reorganization of go-ptop to support comprehensive JVM memory analysis through correlation of multiple data sources: /proc filesystem (smaps), JVM Native Memory Tracking (NMT), and kernel thread information.

## Core Architecture Components

### 1. Memory Reader (`memory/`)
Struct & interface to read smaps and other /proc memory interfaces.

### 2. JVM Client (`jvm/`)
Struct & interface to invoke NMT via jcmd and other JVM Attach API commands.

### 3. Process Info Reader (`proc/`)
Interface to read kernel thread information and I/O statistics.

### 4. Analyzer/Correlator (`analyzer/`)
Interface to correlate memory data from smaps, NMT, and thread information.

### 5. Formatter (`formatter/`)
Output formatter for various formats (table, JSON, text).

### 6. UI (`ui/`)
TUI interface (optional, can also use CLI).

### 7. Configuration (`config/`)
Configuration and options management.

### 8. Error Handling
Standardized error types and graceful degradation.

## Proposed Package Structure

```
go-ptop/
├── main.go                        # Entry point, CLI arg parsing
│
├── memory/
│   ├── smaps.go                   # /proc/pid/smaps reader (Linux)
│   ├── smaps_linux.go             # Linux-specific implementation
│   ├── smaps_darwin.go            # macOS-specific implementation
│   ├── smaps_windows.go           # Windows-specific implementation
│   ├── smaps_types.go             # ProcessMemorySegment, etc.
│   ├── pagemap.go                 # /proc/pid/pagemap (root-only, Linux)
│   ├── numa.go                    # /proc/pid/numa_maps (root-only, Linux)
│   └── memory.go                  # MemoryReader interface
│
├── jvm/
│   ├── attach.go                  # JVM Attach API abstraction
│   ├── attach_linux.go            # Unix socket implementation
│   ├── attach_darwin.go           # macOS (same as Linux)
│   ├── attach_windows.go          # Named pipe implementation
│   ├── nmt.go                     # NMT query & parsing
│   ├── nmt_types.go               # NMTReport, NMTCategory structs
│   ├── threaddump.go              # Thread dump (refactor from jstack.go)
│   ├── jvm_types.go               # JavaThread, etc.
│   └── jvm.go                     # JVMClient interface
│
├── proc/
│   ├── threads.go                 # Kernel thread enumeration
│   ├── io.go                      # Thread I/O stats
│   ├── stat.go                    # /proc/pid/stat parsing
│   └── proc.go                    # ProcReader interface
│
├── analyzer/
│   ├── correlator.go              # Correlate smaps + NMT + threads
│   ├── classifier.go              # Memory purpose classification
│   ├── analyzer_types.go          # TaskMemorySegment, AnalysisResult
│   └── analyzer.go                # Analyzer interface
│
├── formatter/
│   ├── table.go                   # Table output (CLI)
│   ├── json.go                    # JSON output
│   ├── text.go                    # Plain text summary
│   └── formatter.go               # Formatter interface
│
├── ui/
│   ├── tui.go                     # TUI implementation (refactor)
│   ├── tabs.go                    # Tab components
│   ├── sorting.go                 # Sorting strategies
│   └── ui.go                      # UI interface
│
├── config/
│   └── config.go                  # Configuration management
│
├── util/
│   └── format.go                  # String formatting utilities
│
├── pmap.go                        # Existing - to be refactored
├── jstack.go                      # Existing - to be refactored
├── proc.go                        # Existing - may keep parts
├── tui.go                         # Existing - to be refactored
├── utility.go                     # Existing - to be refactored
├── cli.go                         # Existing - empty, can remove
│
├── go.mod
├── go.sum
├── README.md
├── CLAUDE.md
└── DESIGN.md
```

## Interface Definitions

### 1. Memory Reader (`memory/`)

```go
package memory

// MemoryReader reads process memory information from /proc filesystem
type MemoryReader interface {
    // ReadSMaps reads /proc/pid/smaps
    ReadSMaps(pid int32) ([]ProcessMemorySegment, error)

    // ReadPageMap reads /proc/pid/pagemap (requires root)
    ReadPageMap(pid int32, virtAddr, length uint64) ([]PageInfo, error)

    // ReadNUMAMaps reads /proc/pid/numa_maps (requires root)
    ReadNUMAMaps(pid int32) ([]NUMASegment, error)
}

// ProcessMemorySegment represents a memory mapping from smaps
type ProcessMemorySegment struct {
    StartAddr     uint64
    EndAddr       uint64
    Permissions   string
    Offset        uint64
    Device        string
    Inode         uint64
    Path          string

    // Memory stats (from smaps)
    Size          uint64  // KB
    RSS           uint64  // KB
    PSS           uint64  // KB
    SharedClean   uint64
    SharedDirty   uint64
    PrivateClean  uint64
    PrivateDirty  uint64
    Referenced    uint64
    Anonymous     uint64
    Swap          uint64

    // Metadata
    SegmentType   SegmentType  // heap, stack, mmap, etc.
}

type SegmentType string

const (
    SegmentTypeHeap      SegmentType = "heap"
    SegmentTypeStack     SegmentType = "stack"
    SegmentTypeMmap      SegmentType = "mmap"
    SegmentTypeVDSO      SegmentType = "vdso"
    SegmentTypeVVar      SegmentType = "vvar"
    SegmentTypeAnon      SegmentType = "anonymous"
    SegmentTypeUnknown   SegmentType = "unknown"
)

// PageInfo represents page-level information (root-only)
type PageInfo struct {
    VirtualAddr   uint64
    PhysicalPFN   uint64
    IsPresent     bool
    IsSwapped     bool
    IsFileBacked  bool
    Flags         uint64
}
```

### 2. JVM Client (`jvm/`)

```go
package jvm

// JVMClient communicates with JVM via Attach API
type JVMClient interface {
    // Connect establishes connection to JVM
    Connect(pid int32) error

    // Close closes connection
    Close() error

    // GetThreadDump gets full thread dump
    GetThreadDump() (ThreadDump, error)

    // GetNMT gets Native Memory Tracking report
    GetNMT(detail bool) (*NMTReport, error)

    // GetHeapInfo gets heap information
    GetHeapInfo() (*HeapInfo, error)

    // ExecuteCommand executes arbitrary jcmd command
    ExecuteCommand(command string, args ...string) (string, error)
}

// NMTReport represents parsed NMT output
type NMTReport struct {
    Timestamp     time.Time
    Total         NMTCategory
    Categories    map[string]NMTCategory
}

// NMTCategory represents a single NMT category
type NMTCategory struct {
    Name          string
    Reserved      uint64  // KB
    Committed     uint64  // KB

    // Sub-categories (for detail mode)
    Details       []NMTDetail
}

// NMTDetail represents detailed breakdown
type NMTDetail struct {
    Type          string  // malloc, mmap, arena
    Size          uint64  // KB
    Count         int     // Number of allocations
}

// ThreadDump represents parsed thread dump
type ThreadDump struct {
    Timestamp     time.Time
    Threads       []JavaThread
}

// JavaThread represents a Java thread
type JavaThread struct {
    Name          string
    TID           string   // Java thread ID (hex)
    NID           int      // Native thread ID (decimal)
    State         string
    StackPtr      uint64   // Stack pointer address
    Daemon        bool
    Priority      int
}
```

### 3. Proc Reader (`proc/`)

```go
package proc

// ProcReader reads process information from /proc
type ProcReader interface {
    // GetThreads lists all kernel threads
    GetThreads(pid int32) ([]KernelThread, error)

    // GetThreadIOStats gets I/O stats for a thread
    GetThreadIOStats(pid, tid int32) (*IOStats, error)

    // GetThreadStat gets stat info for a thread
    GetThreadStat(pid, tid int32) (*ThreadStat, error)
}

// KernelThread represents a kernel-level thread
type KernelThread struct {
    PID           int32
    TID           int32
    StartStack    uint64   // Stack start address
    State         string
    Name          string
}

// IOStats represents thread I/O statistics
type IOStats struct {
    ReadCount     uint64
    WriteCount    uint64
    ReadBytes     uint64
    WriteBytes    uint64
}

// ThreadStat represents /proc/pid/task/tid/stat data
type ThreadStat struct {
    TID           int32
    Name          string
    State         string
    StartStack    uint64
    // ... other fields
}
```

### 4. Analyzer/Correlator (`analyzer/`)

```go
package analyzer

// Analyzer correlates memory data from multiple sources
type Analyzer interface {
    // Analyze performs full analysis
    Analyze(pid int32) (*AnalysisResult, error)

    // CorrelateThreads correlates Java threads with memory segments
    CorrelateThreads(
        javaThreads []jvm.JavaThread,
        kernelThreads []proc.KernelThread,
        memSegs []memory.ProcessMemorySegment,
    ) ([]ThreadMemorySegment, error)

    // ClassifyMemory classifies memory segments by purpose
    ClassifyMemory(
        memSegs []memory.ProcessMemorySegment,
        nmt *jvm.NMTReport,
    ) ([]ClassifiedSegment, error)
}

// AnalysisResult represents complete analysis
type AnalysisResult struct {
    PID               int32
    Timestamp         time.Time

    // Memory breakdown
    TotalRSS          uint64
    TotalVirtual      uint64

    // Categorized memory
    JavaHeap          MemoryCategory
    Metaspace         MemoryCategory
    CodeCache         MemoryCategory
    ThreadStacks      MemoryCategory
    DirectBuffers     MemoryCategory
    NativeMemory      MemoryCategory
    SharedLibraries   MemoryCategory
    Other             MemoryCategory

    // Thread details
    Threads           []ThreadMemorySegment

    // Raw data
    MemorySegments    []memory.ProcessMemorySegment
    NMTReport         *jvm.NMTReport
}

// ThreadMemorySegment represents a thread's memory
type ThreadMemorySegment struct {
    // Thread info
    JavaThread        *jvm.JavaThread
    KernelThread      proc.KernelThread

    // Memory info
    MemorySegment     memory.ProcessMemorySegment

    // I/O stats
    IOStats           proc.IOStats
}

// ClassifiedSegment represents a memory segment with purpose classification
type ClassifiedSegment struct {
    memory.ProcessMemorySegment
    Purpose           MemoryPurpose
    Confidence        float64  // 0.0 to 1.0
}

type MemoryPurpose string

const (
    PurposeJavaHeap       MemoryPurpose = "Java Heap"
    PurposeMetaspace      MemoryPurpose = "Metaspace"
    PurposeCodeCache      MemoryPurpose = "Code Cache"
    PurposeThreadStack    MemoryPurpose = "Thread Stack"
    PurposeDirectBuffer   MemoryPurpose = "Direct Buffer"
    PurposeNativeLibrary  MemoryPurpose = "Native Library"
    PurposeGCMetadata     MemoryPurpose = "GC Metadata"
    PurposeJITCode        MemoryPurpose = "JIT Code"
    PurposeUnknown        MemoryPurpose = "Unknown"
)

// MemoryCategory represents aggregated memory by category
type MemoryCategory struct {
    Purpose           MemoryPurpose
    TotalSize         uint64
    RSS               uint64
    Committed         uint64  // From NMT
    Segments          []ClassifiedSegment
}
```

### 5. Formatter (`formatter/`)

```go
package formatter

// Formatter formats analysis results for output
type Formatter interface {
    // Format formats the analysis result
    Format(result *analyzer.AnalysisResult) (string, error)
}

// TableFormatter outputs table format
type TableFormatter struct {
    Width         int
    SortBy        SortColumn
    TopN          int
}

// JSONFormatter outputs JSON
type JSONFormatter struct {
    Pretty        bool
}

// TextFormatter outputs plain text summary
type TextFormatter struct {
    Verbose       bool
}

type SortColumn string

const (
    SortByRSS         SortColumn = "rss"
    SortByThreadID    SortColumn = "tid"
    SortByWriteCount  SortColumn = "write_count"
    SortByReadCount   SortColumn = "read_count"
)
```

### 6. UI (`ui/`)

```go
package ui

// UI represents the user interface
type UI interface {
    // Run starts the UI loop
    Run(pid int32) error

    // Stop stops the UI
    Stop()
}

// TUI represents terminal UI
type TUI struct {
    UpdateInterval time.Duration
    Analyzer       analyzer.Analyzer
    // ... termui components
}
```

### 7. Configuration (`config/`)

```go
package config

type Config struct {
    // Target
    PID               int32

    // Features
    EnableNMT         bool
    EnableRootFeatures bool  // pagemap, numa_maps

    // Output
    OutputFormat      string  // "tui", "table", "json", "text"
    UpdateInterval    time.Duration

    // Filtering
    ShowThreads       bool
    ShowMmaps         bool
    TopN              int

    // Sorting
    SortBy            formatter.SortColumn

    // Logging
    Verbose           bool
    LogLevel          string
}
```

### 8. Error Types

```go
package analyzer

type ErrJVMNotFound struct{ PID int32 }
func (e ErrJVMNotFound) Error() string {
    return fmt.Sprintf("JVM process with PID %d not found", e.PID)
}

type ErrNMTNotEnabled struct{}
func (e ErrNMTNotEnabled) Error() string {
    return "Native Memory Tracking (NMT) is not enabled. " +
           "Start JVM with -XX:NativeMemoryTracking=summary or =detail"
}

type ErrRequiresRoot struct{ Feature string }
func (e ErrRequiresRoot) Error() string {
    return fmt.Sprintf("Feature '%s' requires root privileges", e.Feature)
}

type ErrAttachFailed struct{ Reason error }
func (e ErrAttachFailed) Error() string {
    return fmt.Sprintf("Failed to attach to JVM: %v", e.Reason)
}
```

## Migration Path

### Phase 1: Extract Interfaces (No Behavior Change)
- Create package structure
- Define all interfaces
- Move existing code into packages
- Ensure backward compatibility

**Files to refactor:**
- `pmap.go` → `memory/smaps.go`
- `jstack.go` → `jvm/threadump.go` and `jvm/attach.go`
- `proc.go` → `proc/threads.go`, `proc/io.go`, `proc/stat.go`
- `utility.go` → `util/format.go`
- `tui.go` → `ui/tui.go`, `analyzer/correlator.go`

### Phase 2: Add NMT Support
- Implement `jvm/nmt.go`
- Implement NMT parser for summary and detail modes
- Extend analyzer to use NMT data
- Add correlation between NMT categories and smaps regions

### Phase 3: Enhance Correlation
- Improve memory classification algorithm
- Better NMT + smaps correlation
- Add confidence scoring for classifications
- Detect and report discrepancies between NMT and smaps

### Phase 4: Add Root Features (Optional)
- Implement `memory/pagemap.go`
- Add NUMA support via `memory/numa.go`
- Physical page sharing analysis
- Page-level precision tracking

### Phase 5: Multiple Output Formats
- Implement JSON formatter
- Implement CLI table formatter
- Keep TUI as primary interface
- Add summary text formatter

## Key Design Principles

### Separation of Concerns
Each package has a single, well-defined responsibility:
- `memory` - Reads /proc memory info
- `jvm` - Communicates with JVM
- `proc` - Reads /proc process info
- `analyzer` - Correlates and classifies data
- `formatter` - Formats output
- `ui` - User interface

### Interface-Driven Design
All major components are interfaces, enabling:
- Easy testing with mocks
- Multiple implementations
- Clear contracts between components

### Graceful Degradation
Handle missing features gracefully:
- NMT not enabled → Fall back to smaps-only analysis
- No root access → Skip pagemap/numa features
- Attach API fails → Provide helpful error message

### Extensibility
Easy to add:
- New output formats (CSV, HTML, etc.)
- New data sources (perf, eBPF, etc.)
- New analysis algorithms
- New UI modes

## Data Flow

```
┌─────────────┐
│   main.go   │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────────┐
│            Analyzer                     │
│  ┌────────────────────────────────┐    │
│  │ 1. Collect Data                │    │
│  │    ├─ MemoryReader.ReadSMaps() │    │
│  │    ├─ JVMClient.GetNMT()       │    │
│  │    ├─ JVMClient.GetThreadDump()│    │
│  │    └─ ProcReader.GetThreads()  │    │
│  │                                 │    │
│  │ 2. Correlate                    │    │
│  │    ├─ Match Java ↔ Kernel threads  │
│  │    └─ Match threads ↔ memory   │    │
│  │                                 │    │
│  │ 3. Classify                     │    │
│  │    └─ Determine memory purpose │    │
│  │                                 │    │
│  │ 4. Return AnalysisResult        │    │
│  └────────────────────────────────┘    │
└─────────────────┬───────────────────────┘
                  │
                  ▼
         ┌────────────────┐
         │   Formatter    │
         └────────┬───────┘
                  │
                  ▼
         ┌────────────────┐
         │  UI (TUI/CLI)  │
         └────────────────┘
```

## Current Code Mapping

| Current File | New Location(s) |
|--------------|-----------------|
| `main.go` | `cmd/go-ptop/main.go` |
| `pmap.go` | `memory/smaps.go`, `memory/smaps_types.go` |
| `jstack.go` | `jvm/attach.go`, `jvm/threadump.go`, `jvm/jvm_types.go` |
| `proc.go` | `proc/threads.go`, `proc/io.go`, `proc/stat.go` |
| `tui.go` (correlation logic) | `analyzer/correlator.go` |
| `tui.go` (UI logic) | `ui/tui.go`, `ui/tabs.go` |
| `utility.go` | `util/format.go` |
| `cli.go` | (empty, can remove) |

## Implementation Priority

### High Priority (Core Functionality)
1. Phase 1: Package reorganization
2. Phase 2: NMT integration
3. Memory classification improvements

### Medium Priority (Enhanced Features)
4. Phase 5: Alternative output formats (JSON, table)
5. Better error handling and user feedback
6. Configuration management

### Low Priority (Advanced Features)
7. Phase 4: Root-only features (pagemap, NUMA)
8. Performance optimizations
9. Additional JVM metrics (heap info, GC stats)

## Benefits of This Design

### Maintainability
- Clear separation of concerns
- Each component can be tested independently
- Easy to understand code organization

### Extensibility
- Easy to add new data sources
- Easy to add new output formats
- Easy to add new analysis algorithms

### Testability
- Interface-driven design enables mocking
- Small, focused functions
- Clear input/output contracts

### Flexibility
- Can run in multiple modes (TUI, CLI, daemon)
- Can output in multiple formats
- Can gracefully handle missing features

## Future Enhancements

### Additional Data Sources
- eBPF-based allocation tracking
- perf integration for cache misses
- JFR (Java Flight Recorder) integration

### Advanced Analysis
- Memory leak detection
- Allocation hotspot identification
- Historical tracking and trend analysis
- Anomaly detection

### Additional Output Formats
- CSV export
- HTML report with charts
- Prometheus metrics export
- Integration with monitoring systems

### Performance Optimizations
- Incremental updates (only changed segments)
- Parallel data collection
- Caching of expensive operations

## Cross-Platform Feasibility Analysis

### Current Linux Dependencies

The current implementation is **heavily Linux-specific** with the following dependencies:

| Component | Linux-Specific API | Impact |
|-----------|-------------------|---------|
| **Memory Maps** | `/proc/[pid]/smaps` | Critical - core feature |
| **Thread Info** | `/proc/[pid]/task/[tid]/stat` | Critical - thread correlation |
| **Thread I/O** | `/proc/[pid]/task/[tid]/io` | Medium - I/O statistics |
| **JVM Attach** | Unix domain sockets at `/tmp/.java_pid<pid>` | Critical - NMT/thread dumps |
| **Signal Handling** | `unix.SIGQUIT` | Critical - trigger attach listener |
| **File Trigger** | `/proc/[pid]/cwd/.attach_pid<pid>` | Critical - attach protocol |

### Windows Support Feasibility

**Overall Assessment: MEDIUM FEASIBILITY (50-60% feature parity possible)**

#### What's Available on Windows

| Feature | Linux API | Windows Equivalent | Feasibility | Notes |
|---------|-----------|-------------------|-------------|-------|
| **Process Memory** | `/proc/[pid]/smaps` | `VirtualQueryEx()` + `K32GetProcessMemoryInfo()` | ✅ **High** | Can enumerate memory regions, get working set info |
| **Memory Details** | smaps RSS, PSS, etc. | `PSAPI_WORKING_SET_EX_INFORMATION` | ⚠️ **Medium** | Less granular than Linux smaps |
| **Thread List** | `/proc/[pid]/task/*` | `CreateToolhelp32Snapshot()` or `NtQuerySystemInformation()` | ✅ **High** | Full thread enumeration available |
| **Thread Stacks** | `/proc/[pid]/task/[tid]/stat` | `NtQueryInformationThread(ThreadBasicInformation)` | ⚠️ **Medium** | Stack base available, less detail |
| **Thread I/O** | `/proc/[pid]/task/[tid]/io` | `GetThreadIOPendingFlag()` | ❌ **Low** | Windows lacks per-thread I/O stats |
| **JVM Attach API** | Unix domain socket | **Named Pipes** or **TCP loopback** | ⚠️ **Medium** | JVM on Windows uses different attach mechanism |
| **Signal Handling** | `SIGQUIT` | `GenerateConsoleCtrlEvent()` or attach API | ⚠️ **Medium** | Different trigger mechanism |

#### Windows-Specific Challenges

**Critical Blockers:**
1. **JVM Attach API Differences**
   - Windows JVM uses **named pipes** instead of Unix sockets: `\\.\pipe\attach_pid<pid>`
   - Attach protocol differs slightly
   - Requires Windows-specific implementation

2. **No `/proc` Filesystem**
   - Must use Win32 API calls: `VirtualQueryEx()`, `ReadProcessMemory()`
   - Significantly different code paths

3. **Limited Memory Attribution**
   - Windows lacks PSS (Proportional Set Size)
   - Shared memory accounting is less precise
   - Cannot easily distinguish private vs. shared pages

**What Would Work:**
- ✅ Enumerating process memory regions
- ✅ Getting working set size (RSS equivalent)
- ✅ Reading JVM NMT (if attach API works)
- ✅ Thread enumeration and correlation
- ✅ Basic memory categorization

**What Would NOT Work:**
- ❌ Precise PSS calculations
- ❌ Per-thread I/O statistics
- ❌ Fine-grained shared memory analysis
- ❌ Page-level flags (dirty, referenced, etc.)

#### Windows Implementation Path

```go
// memory/smaps_windows.go
package memory

import (
    "golang.org/x/sys/windows"
    "syscall"
    "unsafe"
)

type WindowsMemoryReader struct{}

func (r *WindowsMemoryReader) ReadSMaps(pid int32) ([]ProcessMemorySegment, error) {
    // 1. OpenProcess with PROCESS_QUERY_INFORMATION | PROCESS_VM_READ
    handle, err := windows.OpenProcess(
        windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ,
        false,
        uint32(pid),
    )
    if err != nil {
        return nil, err
    }
    defer windows.CloseHandle(handle)

    // 2. Enumerate memory regions using VirtualQueryEx
    var segments []ProcessMemorySegment
    var address uintptr = 0

    for {
        var memInfo windows.MemoryBasicInformation
        err := windows.VirtualQueryEx(
            handle,
            address,
            &memInfo,
            unsafe.Sizeof(memInfo),
        )
        if err != nil {
            break
        }

        // 3. Convert to ProcessMemorySegment
        segment := ProcessMemorySegment{
            StartAddr: uint64(memInfo.BaseAddress),
            EndAddr:   uint64(memInfo.BaseAddress) + uint64(memInfo.RegionSize),
            Size:      uint64(memInfo.RegionSize) / 1024, // Convert to KB
            // Note: RSS requires additional K32GetWsChanges call
        }

        segments = append(segments, segment)
        address += memInfo.RegionSize
    }

    return segments, nil
}
```

**Estimated Effort:** 3-4 weeks for basic Windows support (without I/O stats)

### macOS Support Feasibility

**Overall Assessment: MEDIUM-HIGH FEASIBILITY (60-70% feature parity possible)**

#### What's Available on macOS

| Feature | Linux API | macOS Equivalent | Feasibility | Notes |
|---------|-----------|------------------|-------------|-------|
| **Process Memory** | `/proc/[pid]/smaps` | `mach_vm_region()` + `proc_pidinfo(PROC_PIDREGIONINFO)` | ✅ **High** | Good memory region enumeration |
| **Memory Details** | smaps RSS, PSS, etc. | `task_info(TASK_VM_INFO)` | ✅ **High** | Excellent detail available |
| **Thread List** | `/proc/[pid]/task/*` | `task_threads()` | ✅ **High** | Full thread enumeration |
| **Thread Stacks** | `/proc/[pid]/task/[tid]/stat` | `thread_info(THREAD_IDENTIFIER_INFO)` | ✅ **High** | Good thread details |
| **Thread I/O** | `/proc/[pid]/task/[tid]/io` | `proc_pidinfo(PROC_PIDTHREADINFO)` | ⚠️ **Medium** | Limited I/O stats per thread |
| **JVM Attach API** | Unix domain socket | **Same as Linux!** | ✅ **High** | macOS JVM uses Unix sockets |
| **Signal Handling** | `SIGQUIT` | **Same as Linux!** | ✅ **High** | POSIX signals work |

#### macOS-Specific Advantages

**What Works Better than Windows:**
1. **JVM Attach API** - Nearly identical to Linux (Unix sockets at `/tmp/.java_pid<pid>`)
2. **Mach VM API** - Very powerful memory introspection
3. **POSIX Signals** - Full signal support
4. **Thread Info** - Rich thread information via Mach APIs

**Challenges:**
1. **No `/proc` Filesystem** - Must use Mach/BSD APIs
2. **Different Memory Model** - macOS uses Mach VM subsystem
3. **SIP (System Integrity Protection)** - May restrict access to some processes

**What Would Work:**
- ✅ Full JVM NMT support (attach API identical)
- ✅ Comprehensive memory region enumeration
- ✅ Thread enumeration and correlation
- ✅ RSS, dirty pages, resident pages
- ✅ Stack pointer correlation
- ✅ Shared memory detection

**What Would Have Limited Support:**
- ⚠️ Per-thread I/O statistics (less detail than Linux)
- ⚠️ PSS calculations (macOS has different accounting)

#### macOS Implementation Path

```go
// memory/smaps_darwin.go
package memory

/*
#include <mach/mach.h>
#include <mach/mach_vm.h>
#include <libproc.h>

// Helper function to get task for pid
kern_return_t get_task_for_pid(int pid, mach_port_t *task) {
    return task_for_pid(mach_task_self(), pid, task);
}
*/
import "C"
import "unsafe"

type DarwinMemoryReader struct{}

func (r *DarwinMemoryReader) ReadSMaps(pid int32) ([]ProcessMemorySegment, error) {
    var task C.mach_port_t

    // 1. Get Mach task port for process
    kr := C.get_task_for_pid(C.int(pid), &task)
    if kr != C.KERN_SUCCESS {
        return nil, fmt.Errorf("task_for_pid failed: %d", kr)
    }
    defer C.mach_port_deallocate(C.mach_task_self(), task)

    // 2. Enumerate VM regions
    var segments []ProcessMemorySegment
    var address C.mach_vm_address_t = 0

    for {
        var size C.mach_vm_size_t
        var info C.vm_region_basic_info_data_64_t
        var count C.mach_msg_type_number_t = C.VM_REGION_BASIC_INFO_COUNT_64
        var objectName C.mach_port_t

        kr := C.mach_vm_region(
            task,
            &address,
            &size,
            C.VM_REGION_BASIC_INFO_64,
            (*C.vm_region_info_t)(unsafe.Pointer(&info)),
            &count,
            &objectName,
        )

        if kr != C.KERN_SUCCESS {
            break
        }

        // 3. Get additional memory info via proc_pidinfo
        var pti C.struct_proc_taskinfo
        C.proc_pidinfo(
            C.int(pid),
            C.PROC_PIDTASKINFO,
            0,
            unsafe.Pointer(&pti),
            C.PROC_PIDTASKINFO_SIZE,
        )

        segment := ProcessMemorySegment{
            StartAddr: uint64(address),
            EndAddr:   uint64(address) + uint64(size),
            Size:      uint64(size) / 1024,
            RSS:       uint64(pti.pti_resident_size) / 1024,
            // Can extract permissions, shared/private, etc.
        }

        segments = append(segments, segment)
        address += C.mach_vm_address_t(size)
    }

    return segments, nil
}
```

**Estimated Effort:** 2-3 weeks for basic macOS support

### Cross-Platform Architecture Recommendations

#### 1. Platform Abstraction Layer

```go
// memory/memory.go
package memory

type MemoryReader interface {
    ReadSMaps(pid int32) ([]ProcessMemorySegment, error)
}

// Factory function
func NewMemoryReader() MemoryReader {
    switch runtime.GOOS {
    case "linux":
        return &LinuxMemoryReader{}
    case "windows":
        return &WindowsMemoryReader{}
    case "darwin":
        return &DarwinMemoryReader{}
    default:
        return nil
    }
}
```

#### 2. Feature Flags

```go
// platform/capabilities.go
package platform

type Capabilities struct {
    SupportsMemoryMaps       bool
    SupportsPSS              bool  // Proportional Set Size
    SupportsThreadIO         bool
    SupportsJVMAttach        bool
    SupportsPageMap          bool  // Root-only
    SupportsNUMA             bool
}

func GetCapabilities() Capabilities {
    switch runtime.GOOS {
    case "linux":
        return Capabilities{
            SupportsMemoryMaps:   true,
            SupportsPSS:          true,
            SupportsThreadIO:     true,
            SupportsJVMAttach:    true,
            SupportsPageMap:      true,
            SupportsNUMA:         true,
        }
    case "darwin":
        return Capabilities{
            SupportsMemoryMaps:   true,
            SupportsPSS:          false,  // Different accounting
            SupportsThreadIO:     true,   // Limited
            SupportsJVMAttach:    true,
            SupportsPageMap:      false,
            SupportsNUMA:         false,
        }
    case "windows":
        return Capabilities{
            SupportsMemoryMaps:   true,
            SupportsPSS:          false,
            SupportsThreadIO:     false,  // Not per-thread
            SupportsJVMAttach:    true,   // Different protocol
            SupportsPageMap:      false,
            SupportsNUMA:         true,   // Via NUMA API
        }
    }
}
```

#### 3. Platform-Specific Build Tags

```go
// memory/smaps_linux.go
//go:build linux

package memory
// Linux implementation

// memory/smaps_windows.go
//go:build windows

package memory
// Windows implementation

// memory/smaps_darwin.go
//go:build darwin

package memory
// macOS implementation
```

### Feature Parity Matrix

| Feature | Linux | macOS | Windows | Priority |
|---------|-------|-------|---------|----------|
| **Memory Maps** | ✅ 100% | ✅ 90% | ⚠️ 70% | **Critical** |
| **RSS/Working Set** | ✅ 100% | ✅ 100% | ✅ 95% | **Critical** |
| **PSS (Proportional Set)** | ✅ 100% | ❌ 0% | ❌ 0% | **High** |
| **Thread Enumeration** | ✅ 100% | ✅ 100% | ✅ 95% | **Critical** |
| **Thread Stack Info** | ✅ 100% | ✅ 90% | ⚠️ 60% | **High** |
| **Thread I/O Stats** | ✅ 100% | ⚠️ 40% | ❌ 0% | **Medium** |
| **JVM NMT via Attach** | ✅ 100% | ✅ 95% | ⚠️ 70% | **Critical** |
| **Memory Classification** | ✅ 100% | ✅ 85% | ⚠️ 60% | **High** |
| **Page-level Analysis** | ✅ 100% (root) | ❌ 0% | ❌ 0% | **Low** |
| **NUMA Info** | ✅ 100% (root) | ❌ 0% | ✅ 80% | **Low** |

### Recommendation: Tiered Platform Support

#### Tier 1: Full Support (Linux)
- All features available
- Primary development platform
- 100% feature parity

#### Tier 2: Good Support (macOS)
- Core features: ✅ Memory analysis, NMT, thread correlation
- Missing: PSS, page-level analysis, NUMA
- **Effort: 2-3 weeks**
- **Benefit: High** (many Java developers use macOS)

#### Tier 3: Basic Support (Windows)
- Core features: ✅ Memory analysis, NMT
- Missing: PSS, thread I/O, fine-grained analysis
- **Effort: 3-4 weeks**
- **Benefit: Medium** (fewer Linux-like use cases on Windows)

### Implementation Priority

**Recommended Approach:**

1. **Phase 1:** Refactor current code with platform abstraction (1 week)
   - Create `MemoryReader` interface
   - Move Linux code to `smaps_linux.go`
   - Add capability detection

2. **Phase 2:** Add macOS support (2-3 weeks)
   - Implement `DarwinMemoryReader`
   - Adapt JVM attach (minimal changes needed)
   - Test on macOS with real JVM

3. **Phase 3:** Add Windows support (3-4 weeks) - **Optional**
   - Implement `WindowsMemoryReader`
   - Implement Windows JVM attach via named pipes
   - Accept limitations (no thread I/O, no PSS)

### Conclusion

- **macOS: RECOMMENDED** - High feasibility, good ROI, minimal effort
- **Windows: OPTIONAL** - Medium feasibility, moderate effort, limited by platform constraints
- **Linux: PRIMARY** - Continue as main platform with full features
