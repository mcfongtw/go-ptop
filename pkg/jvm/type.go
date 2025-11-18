package jvm

import (
	"errors"
	"time"
)

// JVMClient communicates with a target JVM via the Attach API or other transports.
type JVMClient interface {
	Connect(pid int32) error
	Close() error

	GetThreadDump() (*ThreadDump, error)
	GetNMT(detail bool) (*NMTReport, error)
	GetHeapInfo() (*HeapInfo, error)
	ExecuteCommand(command string, args ...string) (string, error)
}

// ThreadDump represents the parsed output of jstack / VM.thread_dump.
type ThreadDump struct {
	Timestamp time.Time
	Threads   []JavaThread
}

// JavaThread captures thread metadata gleaned from the JVM.
type JavaThread struct {
	Name        string
	JavaID      string
	OSID        int32
	Priority    int32
	Daemon      bool
	ThreadState string
	StackTrace  []string
}

// NMTReport represents a parsed Native Memory Tracking report.
type NMTReport struct {
	Timestamp  time.Time
	Summary    []NMTCategory
	Categories []NMTCategory
}

// NMTCategory aggregates memory consumption for a logical JVM component.
type NMTCategory struct {
	Name        string
	ReservedKB  uint64
	CommittedKB uint64
	Details     []NMTCategory
}

// HeapInfo represents high-level heap metrics.
type HeapInfo struct {
	UsedBytes      uint64
	CommittedBytes uint64
	MaxBytes       uint64
	GCType         string
}

var ErrUnsupported = errors.New("JVM attach not supported on this platform")
