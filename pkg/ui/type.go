package ui

import (
	"time"

	"go-ptop/pkg/analyzer"
)

// UI is the top-level interface for all user interaction layers.
type UI interface {
	Run(pid int32) error
	Stop()
}

// TUI defines configuration for the terminal UI implementation.
type TUI struct {
	UpdateInterval time.Duration
	Analyzer       analyzer.Analyzer
}
