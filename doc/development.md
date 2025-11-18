# Development Guide

This document summarizes how to build and run the current MVP of go-ptop locally. Always review [doc/prd/prompt.md](doc/prd/prompt.md) for process guidelines and [doc/prd/requirement.md](doc/prd/requirement.md) for the latest MVP scope.

## Prerequisites
- Go 1.21+ (module uses modern gofmt/go test behavior).
- Linux environment (the MVP only supports Linux due to /proc and JVM attach requirements).
- Target JVM process owned by the same user, with the attach listener enabled (standard HotSpot default).

## Build
```bash
go build -o go-ptop
```
This produces the `go-ptop` binary in the repository root.

## Run
1. Identify the PID of the Java process you want to inspect (e.g., via `jps -l` or `ps -ef | grep java`).
2. Execute go-ptop with the PID:
   ```bash
   ./go-ptop <pid>
   ```
   Example:
   ```bash
   ./go-ptop 12345
   ```
3. The tool attaches to the JVM, parses `/proc/<pid>/smaps`, and prints a table showing kernel/JVM threads with RSS and I/O counters.

### Output Example
```
PID:	12345	Timestamp:	2024-11-14 12:34:56
TID	Thread Name	State	RSS (KB)	Read Cnt	Write Cnt	Read Bytes	Write Bytes
12346	GC Thread#0	runnable	2048	0	0	0	0
...
```

## Error Handling Notes
- If run on a non-Linux platform, the CLI prints an “unsupported” message and exits.
- Attach failures (e.g., insufficient permissions) surface the underlying error (look for `Failed to attach to JVM`).
- Ensure `/tmp/.java_pid<pid>` is accessible; otherwise, go-ptop attempts to trigger it via the standard HotSpot attach protocol.

## Tests
Unit tests exist for the parser packages. Run them with:
```bash
go test ./...
```
This command also serves as a quick sanity check after local edits.

## Docker-based POC (usable from macOS)
When you need a Linux environment on macOS, the repository provides a Docker-based proof-of-concept flow. It builds a container that runs both a sample JVM and go-ptop in the same namespace.

```bash
./build/poc.sh
```

The script:
1. Builds the image described in `build/Dockerfile` (installs Go + OpenJDK, compiles `build/HelloWorld.java` into `helloworld.jar`, builds go-ptop). The script runs `docker build --no-cache ...` to avoid reusing layers with stale `go.mod` versions.
2. Runs the container, which starts `java -jar helloworld.jar`, waits a few seconds, then executes `go-ptop <pid>` against that JVM and prints the table output before shutting down.
