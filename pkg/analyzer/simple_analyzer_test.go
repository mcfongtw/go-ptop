//go:build unit

package analyzer

import (
	"testing"

	"go-ptop/pkg/jvm"
	"go-ptop/pkg/memory"
	"go-ptop/pkg/proc"
)

func Test_ClassifyMemory_UsesSegmentType(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name            string
		segmentType     memory.SegmentType
		expectedPurpose MemoryPurpose
	}{
		{"heap_to_java_heap", memory.SegmentTypeHeap, PurposeJavaHeap},
		{"stack_to_thread_stack", memory.SegmentTypeStack, PurposeThreadStack},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			a := &simpleAnalyzer{}
			segments := []memory.ProcessMemorySegment{
				{SegmentType: tc.segmentType, Size: 100, RSS: 50},
			}

			classified, err := a.ClassifyMemory(segments, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if classified[0].Purpose != tc.expectedPurpose {
				t.Errorf("expected %v, got %v", tc.expectedPurpose, classified[0].Purpose)
			}
		})
	}
}

func Test_CorrelateThreads_MatchesJavaThreads(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name              string
		javaThreadName    string
		javaThreadOSID    int32
		kernelTID         int32
		expectedCorrelated int
	}{
		{
			name:              "worker_thread",
			javaThreadName:    "worker",
			javaThreadOSID:    123,
			kernelTID:         123,
			expectedCorrelated: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			stubProc := &stubProcReader{
				stats: map[int32]*proc.IOStats{
					tc.kernelTID: {ReadCount: 1, WriteCount: 2},
				},
			}

			a := &simpleAnalyzer{procReader: stubProc}
			javaThreads := []jvm.JavaThread{{Name: tc.javaThreadName, OSID: tc.javaThreadOSID}}
			kernelThreads := []proc.KernelThread{{PID: 42, TID: tc.kernelTID, StartStack: 0x2000}}
			segments := []memory.ProcessMemorySegment{{StartAddr: 0x1000, EndAddr: 0x3000, SegmentType: memory.SegmentTypeStack}}

			correlated, err := a.CorrelateThreads(javaThreads, kernelThreads, segments)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(correlated) != tc.expectedCorrelated {
				t.Fatalf("expected %d correlated thread(s), got %d", tc.expectedCorrelated, len(correlated))
			}

			entry := correlated[0]
			if entry.JavaThread == nil || entry.JavaThread.Name != tc.javaThreadName {
				t.Errorf("java thread not linked: %+v", entry.JavaThread)
			}
			if entry.Segment.SegmentType != memory.SegmentTypeStack {
				t.Error("segment not attached")
			}
			if entry.IOStats.ReadCount != 1 {
				t.Error("io stats not populated")
			}
		})
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
