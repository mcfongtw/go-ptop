package formatter

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"text/tabwriter"

	"go-ptop/pkg/analyzer"
)

// NewTableFormatter returns a stub table formatter implementation.
func NewTableFormatter() *TableFormatter {
	return &TableFormatter{
		SortBy: SortByRSS,
		TopN:   50,
		Width:  4,
	}
}

// NewJSONFormatter returns a stub JSON formatter implementation.
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// NewTextFormatter returns a stub text formatter implementation.
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{}
}

var errFormatterNotImplemented = errors.New("formatter not implemented")

// Format renders the analyzer output as a tabular summary for the CLI.
func (f *TableFormatter) Format(result *analyzer.AnalysisResult) (string, error) {
	if result == nil {
		return "", fmt.Errorf("nil analysis result")
	}

	threads := make([]analyzer.ThreadMemorySegment, len(result.Threads))
	copy(threads, result.Threads)

	sort.SliceStable(threads, func(i, j int) bool {
		switch f.SortBy {
		case SortByThreadID:
			return threads[i].KernelThread.TID < threads[j].KernelThread.TID
		case SortByWriteCount:
			return threads[i].IOStats.WriteCount > threads[j].IOStats.WriteCount
		case SortByReadCount:
			return threads[i].IOStats.ReadCount > threads[j].IOStats.ReadCount
		case SortByRSS:
			fallthrough
		default:
			return threads[i].Segment.RSS > threads[j].Segment.RSS
		}
	})

	if f.TopN > 0 && len(threads) > f.TopN {
		threads = threads[:f.TopN]
	}

	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 4, 2, ' ', 0)

	fmt.Fprintf(tw, "PID:\t%d\tTimestamp:\t%s\n", result.PID, result.Timestamp.Format(timeLayout))
	fmt.Fprintln(tw, "TID\tThread Name\tState\tRSS (KB)\tRead Cnt\tWrite Cnt\tRead Bytes\tWrite Bytes")

	for _, entry := range threads {
		name := "<unknown>"
		state := entry.KernelThread.State
		if entry.JavaThread != nil {
			if entry.JavaThread.Name != "" {
				name = entry.JavaThread.Name
			}
			if entry.JavaThread.ThreadState != "" {
				state = entry.JavaThread.ThreadState
			}
		}
		fmt.Fprintf(
			tw,
			"%d\t%s\t%s\t%d\t%d\t%d\t%d\t%d\n",
			entry.KernelThread.TID,
			name,
			state,
			entry.Segment.RSS,
			entry.IOStats.ReadCount,
			entry.IOStats.WriteCount,
			entry.IOStats.ReadBytes,
			entry.IOStats.WriteBytes,
		)
	}

	if err := tw.Flush(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Format outputs JSON but currently only signals lack of implementation.
func (f *JSONFormatter) Format(result *analyzer.AnalysisResult) (string, error) {
	return "", errFormatterNotImplemented
}

// Format outputs text but currently only signals lack of implementation.
func (f *TextFormatter) Format(result *analyzer.AnalysisResult) (string, error) {
	return "", errFormatterNotImplemented
}

const timeLayout = "2006-01-02 15:04:05"
