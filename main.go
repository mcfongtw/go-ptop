package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"go-ptop/pkg/analyzer"
	"go-ptop/pkg/formatter"
	"go-ptop/pkg/jvm"
	"go-ptop/pkg/memory"
	"go-ptop/pkg/proc"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: go-ptop <pid>")
		os.Exit(1)
	}

	pid64, err := strconv.ParseInt(os.Args[1], 10, 32)
	if err != nil || pid64 <= 0 {
		fmt.Fprintf(os.Stderr, "Invalid PID: %v\n", os.Args[1])
		os.Exit(1)
	}
	pid := int32(pid64)

	memReader := memory.NewMemoryReader()
	jvmClient := jvm.NewClient()
	procReader := proc.NewReader()
	an := analyzer.New(memReader, jvmClient, procReader)

	result, err := an.Analyze(pid)
	if err != nil {
		printAnalyzeError(pid, err)
		os.Exit(1)
	}

	table := formatter.NewTableFormatter()
	output, err := table.Format(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(output)
}

func printAnalyzeError(pid int32, err error) {
	switch {
	case errors.Is(err, memory.ErrUnsupported):
		fmt.Fprintln(os.Stderr, "Memory analysis is not supported on this platform. Linux is required for the MVP.")
	case errors.Is(err, jvm.ErrUnsupported):
		fmt.Fprintln(os.Stderr, "JVM attach is not supported on this platform.")
	case errors.Is(err, proc.ErrUnsupported):
		fmt.Fprintln(os.Stderr, "Kernel thread inspection is not supported on this platform.")
	default:
		var attachErr analyzer.ErrAttachFailed
		if errors.As(err, &attachErr) {
			fmt.Fprintf(os.Stderr, "Failed to attach to JVM (pid %d): %v\n", pid, attachErr.Reason)
			return
		}
		fmt.Fprintf(os.Stderr, "Failed to analyze pid %d: %v\n", pid, err)
	}
}
