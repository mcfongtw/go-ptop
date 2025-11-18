package analyzer

import (
	"fmt"
	"time"

	"go-ptop/pkg/jvm"
	"go-ptop/pkg/memory"
	"go-ptop/pkg/proc"
)

// Analyzer correlates information from all data sources.
type Analyzer interface {
	Analyze(pid int32) (*AnalysisResult, error)
	CorrelateThreads(javaThreads []jvm.JavaThread, kernelThreads []proc.KernelThread, segments []memory.ProcessMemorySegment) ([]ThreadMemorySegment, error)
	ClassifyMemory(segments []memory.ProcessMemorySegment, report *jvm.NMTReport) ([]ClassifiedSegment, error)
}

// AnalysisResult stores the merged data returned to UI/formatters.
type AnalysisResult struct {
	PID        int32
	Timestamp  time.Time
	TotalRSS   uint64
	TotalVSize uint64

	JavaHeap        MemoryCategory
	Metaspace       MemoryCategory
	CodeCache       MemoryCategory
	ThreadStacks    MemoryCategory
	DirectBuffers   MemoryCategory
	NativeMemory    MemoryCategory
	SharedLibraries MemoryCategory
	Other           MemoryCategory

	Threads        []ThreadMemorySegment
	MemorySegments []memory.ProcessMemorySegment
	NMTReport      *jvm.NMTReport
}

// ThreadMemorySegment associates thread info with a memory segment and IO stats.
type ThreadMemorySegment struct {
	JavaThread   *jvm.JavaThread
	KernelThread proc.KernelThread
	Segment      memory.ProcessMemorySegment
	IOStats      proc.IOStats
}

// ClassifiedSegment describes a memory mapping with a heuristic purpose.
type ClassifiedSegment struct {
	Segment    memory.ProcessMemorySegment
	Purpose    MemoryPurpose
	Confidence float64
}

// MemoryPurpose enumerates memory uses.
type MemoryPurpose string

const (
	PurposeJavaHeap      MemoryPurpose = "Java Heap"
	PurposeMetaspace     MemoryPurpose = "Metaspace"
	PurposeCodeCache     MemoryPurpose = "Code Cache"
	PurposeThreadStack   MemoryPurpose = "Thread Stack"
	PurposeDirectBuffer  MemoryPurpose = "Direct Buffer"
	PurposeNativeLibrary MemoryPurpose = "Native Library"
	PurposeGCMetadata    MemoryPurpose = "GC Metadata"
	PurposeJITCode       MemoryPurpose = "JIT Code"
	PurposeUnknown       MemoryPurpose = "Unknown"
)

// MemoryCategory aggregates segments by purpose.
type MemoryCategory struct {
	Purpose   MemoryPurpose
	TotalSize uint64
	RSS       uint64
	Committed uint64
	Segments  []ClassifiedSegment
}

// Custom error helpers for common failure scenarios.
type ErrJVMNotFound struct{ PID int32 }

func (e ErrJVMNotFound) Error() string {
	return fmt.Sprintf("JVM process with PID %d not found", e.PID)
}

type ErrNMTNotEnabled struct{}

func (ErrNMTNotEnabled) Error() string {
	return "Native Memory Tracking (NMT) is not enabled"
}

type ErrRequiresRoot struct{ Feature string }

func (e ErrRequiresRoot) Error() string {
	return fmt.Sprintf("Feature '%s' requires root privileges", e.Feature)
}

type ErrAttachFailed struct{ Reason error }

func (e ErrAttachFailed) Error() string {
	return fmt.Sprintf("Failed to attach to JVM: %v", e.Reason)
}
