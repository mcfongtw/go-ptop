# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

go-ptop is a Java process analyzer that provides a terminal-based interface for monitoring Java thread and memory usage. It connects to Java processes via JVM attach API to display real-time thread and memory information in a tabular format.

## Architecture

The codebase is organized into single-file modules:

- **main.go**: Entry point and CLI argument parsing
- **tui.go**: Terminal UI implementation using termui library, handles tabbed display and keyboard interactions
- **jstack.go**: Java thread dump functionality via JVM attach API using Unix domain sockets
- **pmap.go**: Process memory mapping analysis by parsing /proc/pid/smaps
- **proc.go**: Kernel thread information extraction from /proc filesystem
- **utility.go**: String formatting utilities for memory addresses and numeric values

### Key Data Structures

- `TaskMemorySegment`: Combines process memory segments with thread-specific I/O statistics
- `ProcessMemorySegment`: Memory mapping information from /proc/pid/smaps
- `JavaThread`: Java thread metadata from jstack output
- `KernelThread`: Kernel thread information from /proc/pid/task

### Core Workflow

1. Parse target Java process PID from command line
2. Trigger Java thread dump via JVM attach API (SIGQUIT signal)
3. Parse thread information and correlate with kernel thread data
4. Read process memory mappings from /proc/pid/smaps
5. Associate Java threads with memory segments using stack pointers
6. Display real-time data in tabbed TUI with sorting capabilities

## Development Commands

### Build and Run
```bash
# Initialize Go module (if not already done)
go mod init go-ptop

# Download dependencies
go mod tidy

# Build application
go build -o go-ptop

# Run application (requires Java process PID)
./go-ptop <pid>
```

### Dependencies
The project requires these Go modules:
- `github.com/gizak/termui@v2.3.0` - Terminal UI library
- `github.com/golang/glog` - Logging
- `github.com/shirou/gopsutil` - Process utilities
- `golang.org/x/sys/unix` - Unix system calls

### Testing
```bash
go test ./...
```

## Platform Requirements

This tool is designed for Linux systems and requires:
- Java processes running with JVM attach API enabled
- Access to /proc filesystem for process information
- Unix domain socket support for JVM communication

## TUI Controls

- `<Esc>`: Quit application
- `<Left>/<Right>`: Switch between tabs (Thread, MMap, Others, All)
- `<Ctrl-s>`: Sort by Write Count
- `<Ctrl-d>`: Sort by Task ID

## Common Issues

The codebase has some compatibility issues:
- Uses older termui v2.3.0 API
- SIGQUIT constant may need platform-specific handling
- Some struct field references may be outdated with newer library versions

When building, ensure proper Go module initialization and use compatible dependency versions as specified in go.mod.