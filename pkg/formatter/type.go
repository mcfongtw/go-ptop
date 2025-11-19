package formatter

import "go-ptop/pkg/analyzer"

// Formatter renders analysis results for a given output format.
type Formatter interface {
	Format(result *analyzer.AnalysisResult) (string, error)
}

// SortColumn enumerates supported column sorts for table outputs.
type SortColumn string

const (
	SortByRSS        SortColumn = "rss"
	SortByThreadID   SortColumn = "tid"
	SortByWriteCount SortColumn = "write_count"
	SortByReadCount  SortColumn = "read_count"
)

// TableFormatter configuration placeholder.
type TableFormatter struct {
	Width  int
	SortBy SortColumn
	TopN   int
}

// JSONFormatter configuration placeholder.
type JSONFormatter struct {
	Pretty bool
}

// TextFormatter configuration placeholder.
type TextFormatter struct {
	Verbose bool
}
