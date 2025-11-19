package config

import (
	"time"

	"go-ptop/pkg/formatter"
)

// Config centralizes runtime options for go-ptop.
type Config struct {
	PID                int32
	EnableNMT          bool
	EnableRootFeatures bool

	OutputFormat   string
	UpdateInterval time.Duration

	ShowThreads bool
	ShowMMaps   bool
	TopN        int

	SortBy formatter.SortColumn

	Verbose  bool
	LogLevel string
}
