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

func Test_NewTableFormatter_ReturnsDefaults(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name          string
		checkField    string
		expectedValue interface{}
	}{
		{"default_sort_by", "SortBy", SortByRSS},
		{"default_top_n", "TopN", 50},
		{"default_width", "Width", 4},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			f := NewTableFormatter()
			if f == nil {
				t.Fatal("expected non-nil formatter")
			}

			switch tc.checkField {
			case "SortBy":
				if f.SortBy != tc.expectedValue.(SortColumn) {
					t.Errorf("expected SortBy to be %s, got %s", tc.expectedValue, f.SortBy)
				}
			case "TopN":
				if f.TopN != tc.expectedValue.(int) {
					t.Errorf("expected TopN to be %d, got %d", tc.expectedValue, f.TopN)
				}
			case "Width":
				if f.Width != tc.expectedValue.(int) {
					t.Errorf("expected Width to be %d, got %d", tc.expectedValue, f.Width)
				}
			}
		})
	}
}

func Test_TableFormatter_Format_ErrorCases(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name           string
		result         *analyzer.AnalysisResult
		expectError    bool
		errorContains  string
	}{
		{
			name:          "nil_result",
			result:        nil,
			expectError:   true,
			errorContains: "nil",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			f := NewTableFormatter()
			_, err := f.Format(tc.result)

			if tc.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("expected error to contain '%s', got: %s", tc.errorContains, err.Error())
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_TableFormatter_Format_ValidResults(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name            string
		result          *analyzer.AnalysisResult
		expectedStrings []string
	}{
		{
			name: "empty_result",
			result: &analyzer.AnalysisResult{
				PID:       1234,
				Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				Threads:   []analyzer.ThreadMemorySegment{},
			},
			expectedStrings: []string{"1234", "TID"},
		},
		{
			name: "with_threads",
			result: &analyzer.AnalysisResult{
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
			},
			expectedStrings: []string{"main", "100", "RUNNABLE"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			f := NewTableFormatter()
			output, err := f.Format(tc.result)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, expected := range tc.expectedStrings {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain '%s'", expected)
				}
			}
		})
	}
}

func Test_TableFormatter_SortOptions(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name       string
		sortBy     SortColumn
		threads    []analyzer.ThreadMemorySegment
		firstTID   string
		secondTID  string
	}{
		{
			name:   "sort_by_thread_id",
			sortBy: SortByThreadID,
			threads: []analyzer.ThreadMemorySegment{
				{KernelThread: proc.KernelThread{TID: 300}, Segment: memory.ProcessMemorySegment{RSS: 100}},
				{KernelThread: proc.KernelThread{TID: 100}, Segment: memory.ProcessMemorySegment{RSS: 200}},
				{KernelThread: proc.KernelThread{TID: 200}, Segment: memory.ProcessMemorySegment{RSS: 150}},
			},
			firstTID:  "100",
			secondTID: "200",
		},
		{
			name:   "sort_by_write_count",
			sortBy: SortByWriteCount,
			threads: []analyzer.ThreadMemorySegment{
				{KernelThread: proc.KernelThread{TID: 101}, Segment: memory.ProcessMemorySegment{RSS: 50}, IOStats: proc.IOStats{WriteCount: 10}},
				{KernelThread: proc.KernelThread{TID: 202}, Segment: memory.ProcessMemorySegment{RSS: 50}, IOStats: proc.IOStats{WriteCount: 999}},
			},
			firstTID:  "202",
			secondTID: "101",
		},
		{
			name:   "sort_by_read_count",
			sortBy: SortByReadCount,
			threads: []analyzer.ThreadMemorySegment{
				{KernelThread: proc.KernelThread{TID: 103}, Segment: memory.ProcessMemorySegment{RSS: 50}, IOStats: proc.IOStats{ReadCount: 5}},
				{KernelThread: proc.KernelThread{TID: 204}, Segment: memory.ProcessMemorySegment{RSS: 50}, IOStats: proc.IOStats{ReadCount: 888}},
			},
			firstTID:  "204",
			secondTID: "103",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			f := NewTableFormatter()
			f.SortBy = tc.sortBy

			result := &analyzer.AnalysisResult{
				PID:       1234,
				Timestamp: time.Now(),
				Threads:   tc.threads,
			}

			output, err := f.Format(result)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			idxFirst := strings.Index(output, tc.firstTID)
			idxSecond := strings.Index(output, tc.secondTID)

			if idxFirst > idxSecond {
				t.Errorf("expected TID %s to appear before TID %s", tc.firstTID, tc.secondTID)
			}
		})
	}
}

func Test_TableFormatter_TopNLimit(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name      string
		topN      int
		threads   []analyzer.ThreadMemorySegment
		maxLines  int
	}{
		{
			name: "limit_to_two",
			topN: 2,
			threads: []analyzer.ThreadMemorySegment{
				{KernelThread: proc.KernelThread{TID: 1}, Segment: memory.ProcessMemorySegment{RSS: 300}},
				{KernelThread: proc.KernelThread{TID: 2}, Segment: memory.ProcessMemorySegment{RSS: 200}},
				{KernelThread: proc.KernelThread{TID: 3}, Segment: memory.ProcessMemorySegment{RSS: 100}},
			},
			maxLines: 4, // header line + 2 data lines + trailing
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			f := NewTableFormatter()
			f.TopN = tc.topN

			result := &analyzer.AnalysisResult{
				PID:       1234,
				Timestamp: time.Now(),
				Threads:   tc.threads,
			}

			output, err := f.Format(result)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if strings.Count(output, "\n") > tc.maxLines {
				t.Error("expected output to be limited by TopN")
			}
		})
	}
}

func Test_StubFormatters_ReturnNotImplemented(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name      string
		formatter Formatter
	}{
		{"json_formatter", NewJSONFormatter()},
		{"text_formatter", NewTextFormatter()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			result := &analyzer.AnalysisResult{
				PID:       1234,
				Timestamp: time.Now(),
			}

			_, err := tc.formatter.Format(result)
			if err == nil {
				t.Fatal("expected error from stub formatter")
			}
			if !strings.Contains(err.Error(), "not implemented") {
				t.Errorf("expected 'not implemented' error, got: %s", err.Error())
			}
		})
	}
}
