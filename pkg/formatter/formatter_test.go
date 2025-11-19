//go:build unit

package formatter

import (
	"strings"
	"testing"
	"time"

	"go-ptop/pkg/analyzer"
	"go-ptop/pkg/jvm"
	"go-ptop/pkg/memory"
	"go-ptop/pkg/proc"
)

func TestNewTableFormatterReturnsDefaults(t *testing.T) {
	f := NewTableFormatter()

	if f == nil {
		t.Fatal("expected non-nil formatter")
	}
	if f.SortBy != SortByRSS {
		t.Errorf("expected SortBy to be %s, got %s", SortByRSS, f.SortBy)
	}
	if f.TopN != 50 {
		t.Errorf("expected TopN to be 50, got %d", f.TopN)
	}
	if f.Width != 4 {
		t.Errorf("expected Width to be 4, got %d", f.Width)
	}
}

func TestTableFormatterFormatNilResult(t *testing.T) {
	f := NewTableFormatter()

	_, err := f.Format(nil)
	if err == nil {
		t.Fatal("expected error for nil result")
	}
	if !strings.Contains(err.Error(), "nil") {
		t.Errorf("expected error to mention nil, got: %s", err.Error())
	}
}

func TestTableFormatterFormatEmptyResult(t *testing.T) {
	f := NewTableFormatter()

	result := &analyzer.AnalysisResult{
		PID:       1234,
		Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		Threads:   []analyzer.ThreadMemorySegment{},
	}

	output, err := f.Format(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "1234") {
		t.Error("expected output to contain PID")
	}
	if !strings.Contains(output, "TID") {
		t.Error("expected output to contain header")
	}
}

func TestTableFormatterFormatWithThreads(t *testing.T) {
	f := NewTableFormatter()

	result := &analyzer.AnalysisResult{
		PID:       5678,
		Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		Threads: []analyzer.ThreadMemorySegment{
			{
				JavaThread: &jvm.JavaThread{
					Name:        "main",
					ThreadState: "RUNNABLE",
				},
				KernelThread: proc.KernelThread{
					TID:   100,
					State: "R",
				},
				Segment: memory.ProcessMemorySegment{
					RSS: 1024,
				},
				IOStats: proc.IOStats{
					ReadCount:  10,
					WriteCount: 20,
					ReadBytes:  1000,
					WriteBytes: 2000,
				},
			},
		},
	}

	output, err := f.Format(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "main") {
		t.Error("expected output to contain thread name 'main'")
	}
	if !strings.Contains(output, "100") {
		t.Error("expected output to contain TID 100")
	}
	if !strings.Contains(output, "RUNNABLE") {
		t.Error("expected output to contain thread state")
	}
}

func TestTableFormatterSortByThreadID(t *testing.T) {
	f := NewTableFormatter()
	f.SortBy = SortByThreadID

	result := &analyzer.AnalysisResult{
		PID:       1234,
		Timestamp: time.Now(),
		Threads: []analyzer.ThreadMemorySegment{
			{
				KernelThread: proc.KernelThread{TID: 300},
				Segment:      memory.ProcessMemorySegment{RSS: 100},
			},
			{
				KernelThread: proc.KernelThread{TID: 100},
				Segment:      memory.ProcessMemorySegment{RSS: 200},
			},
			{
				KernelThread: proc.KernelThread{TID: 200},
				Segment:      memory.ProcessMemorySegment{RSS: 150},
			},
		},
	}

	output, err := f.Format(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that TID 100 appears before TID 200 which appears before TID 300
	idx100 := strings.Index(output, "100")
	idx200 := strings.Index(output, "200")
	idx300 := strings.Index(output, "300")

	if idx100 > idx200 || idx200 > idx300 {
		t.Error("expected threads to be sorted by TID ascending")
	}
}

func TestTableFormatterTopNLimit(t *testing.T) {
	f := NewTableFormatter()
	f.TopN = 2

	result := &analyzer.AnalysisResult{
		PID:       1234,
		Timestamp: time.Now(),
		Threads: []analyzer.ThreadMemorySegment{
			{KernelThread: proc.KernelThread{TID: 1}, Segment: memory.ProcessMemorySegment{RSS: 300}},
			{KernelThread: proc.KernelThread{TID: 2}, Segment: memory.ProcessMemorySegment{RSS: 200}},
			{KernelThread: proc.KernelThread{TID: 3}, Segment: memory.ProcessMemorySegment{RSS: 100}},
		},
	}

	output, err := f.Format(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only contain TID 1 and 2 (highest RSS), not TID 3
	if strings.Count(output, "\n") > 4 { // header line + 2 data lines + trailing
		t.Error("expected output to be limited by TopN")
	}
}

func TestJSONFormatterReturnsNotImplemented(t *testing.T) {
	f := NewJSONFormatter()

	result := &analyzer.AnalysisResult{
		PID:       1234,
		Timestamp: time.Now(),
	}

	_, err := f.Format(result)
	if err == nil {
		t.Fatal("expected error from stub JSON formatter")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("expected 'not implemented' error, got: %s", err.Error())
	}
}

func TestTextFormatterReturnsNotImplemented(t *testing.T) {
	f := NewTextFormatter()

	result := &analyzer.AnalysisResult{
		PID:       1234,
		Timestamp: time.Now(),
	}

	_, err := f.Format(result)
	if err == nil {
		t.Fatal("expected error from stub text formatter")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("expected 'not implemented' error, got: %s", err.Error())
	}
}

func TestTableFormatterSortByWriteCount(t *testing.T) {
	f := NewTableFormatter()
	f.SortBy = SortByWriteCount

	result := &analyzer.AnalysisResult{
		PID:       1234,
		Timestamp: time.Now(),
		Threads: []analyzer.ThreadMemorySegment{
			{
				KernelThread: proc.KernelThread{TID: 1},
				Segment:      memory.ProcessMemorySegment{RSS: 100},
				IOStats:      proc.IOStats{WriteCount: 10},
			},
			{
				KernelThread: proc.KernelThread{TID: 2},
				Segment:      memory.ProcessMemorySegment{RSS: 100},
				IOStats:      proc.IOStats{WriteCount: 100},
			},
		},
	}

	output, err := f.Format(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// TID 2 should appear before TID 1 (higher write count)
	idx1 := strings.Index(output, "\t1\t")
	idx2 := strings.Index(output, "\t2\t")

	if idx2 > idx1 && idx1 != -1 {
		t.Error("expected thread with higher write count to appear first")
	}
}

func TestTableFormatterSortByReadCount(t *testing.T) {
	f := NewTableFormatter()
	f.SortBy = SortByReadCount

	result := &analyzer.AnalysisResult{
		PID:       1234,
		Timestamp: time.Now(),
		Threads: []analyzer.ThreadMemorySegment{
			{
				KernelThread: proc.KernelThread{TID: 1},
				Segment:      memory.ProcessMemorySegment{RSS: 100},
				IOStats:      proc.IOStats{ReadCount: 5},
			},
			{
				KernelThread: proc.KernelThread{TID: 2},
				Segment:      memory.ProcessMemorySegment{RSS: 100},
				IOStats:      proc.IOStats{ReadCount: 50},
			},
		},
	}

	output, err := f.Format(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// TID 2 should appear before TID 1 (higher read count)
	idx1 := strings.Index(output, "\t1\t")
	idx2 := strings.Index(output, "\t2\t")

	if idx2 > idx1 && idx1 != -1 {
		t.Error("expected thread with higher read count to appear first")
	}
}
