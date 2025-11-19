package proc

import "errors"

// ProcReader abstracts /proc data collection.
type ProcReader interface {
	Threads(pid int32) ([]KernelThread, error)
	ThreadStat(pid int32, tid int32) (*ThreadStat, error)
	ThreadIO(pid int32, tid int32) (*IOStats, error)
}

// KernelThread represents aggregate information about a kernel-level thread.
type KernelThread struct {
	PID        int32
	TID        int32
	StartStack uint64
	State      string
	Name       string
}

// IOStats captures per-thread IO counters.
type IOStats struct {
	ReadCount  uint64
	WriteCount uint64
	ReadBytes  uint64
	WriteBytes uint64
}

// ThreadStat mirrors /proc/<pid>/task/<tid>/stat.
type ThreadStat struct {
	TID        int32
	Name       string
	State      string
	StartStack uint64
}

var ErrUnsupported = errors.New("proc reader not supported on this platform")
