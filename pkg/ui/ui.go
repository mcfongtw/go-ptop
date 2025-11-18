package ui

import (
	"errors"
	"time"

	"go-ptop/pkg/analyzer"
)

var errTUINotImplemented = errors.New("terminal UI not implemented")

// NewTUI returns a stub TUI instance to be wired by main until real UI moves.
func NewTUI(an analyzer.Analyzer) UI {
	return &TUI{
		UpdateInterval: time.Second,
		Analyzer:       an,
	}
}

// Run currently returns a not implemented error as a placeholder.
func (t *TUI) Run(pid int32) error {
	return errTUINotImplemented
}

// Stop is a no-op placeholder.
func (t *TUI) Stop() {}
