//go:build unit

package analyzer

import (
	"testing"

	"go-ptop/pkg/jvm"
	"go-ptop/pkg/memory"
	"go-ptop/pkg/proc"
)

func TestClassifyMemoryUsesSegmentType(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	a := &simpleAnalyzer{}
	segments := []memory.ProcessMemorySegment{
		{SegmentType: memory.SegmentTypeHeap, Size: 100, RSS: 50},
		{SegmentType: memory.SegmentTypeStack, Size: 10, RSS: 5},
	}

	classified, err := a.ClassifyMemory(segments, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if classified[0].Purpose != PurposeJavaHeap {
		t.Fatalf("expected heap purpose, got %v", classified[0].Purpose)
	}
	if classified[1].Purpose != PurposeThreadStack {
		t.Fatalf("expected stack purpose, got %v", classified[1].Purpose)
	}
}

func TestCorrelateThreadsMatchesJavaThreads(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	stubProc := &stubProcReader{
		stats: map[int32]*proc.IOStats{
			123: {ReadCount: 1, WriteCount: 2},
		},
	}

	a := &simpleAnalyzer{procReader: stubProc}
	javaThreads := []jvm.JavaThread{{Name: "worker", OSID: 123}}
	kernelThreads := []proc.KernelThread{{PID: 42, TID: 123, StartStack: 0x2000}}
	segments := []memory.ProcessMemorySegment{{StartAddr: 0x1000, EndAddr: 0x3000, SegmentType: memory.SegmentTypeStack}}

	correlated, err := a.CorrelateThreads(javaThreads, kernelThreads, segments)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(correlated) != 1 {
		t.Fatalf("expected 1 correlated thread, got %d", len(correlated))
	}

	entry := correlated[0]
	if entry.JavaThread == nil || entry.JavaThread.Name != "worker" {
		t.Fatalf("java thread not linked: %+v", entry.JavaThread)
	}
	if entry.Segment.SegmentType != memory.SegmentTypeStack {
		t.Fatalf("segment not attached")
	}
	if entry.IOStats.ReadCount != 1 {
		t.Fatalf("io stats not populated")
	}
}

type stubProcReader struct {
	stats map[int32]*proc.IOStats
}

func (s *stubProcReader) Threads(pid int32) ([]proc.KernelThread, error) {
	return nil, nil
}

func (s *stubProcReader) ThreadStat(pid int32, tid int32) (*proc.ThreadStat, error) {
	return nil, nil
}

func (s *stubProcReader) ThreadIO(pid int32, tid int32) (*proc.IOStats, error) {
	if stat, ok := s.stats[tid]; ok {
		return stat, nil
	}
	return &proc.IOStats{}, nil
}
