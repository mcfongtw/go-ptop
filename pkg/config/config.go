package config

import (
	"time"

	"go-ptop/pkg/formatter"
)

// Default returns the baseline configuration used by the CLI/TUI entry point.
func Default() Config {
	return Config{
		OutputFormat:   "tui",
		UpdateInterval: 5 * time.Second,
		ShowThreads:    true,
		ShowMMaps:      true,
		TopN:           50,
		SortBy:         formatter.SortByRSS,
	}
}

// Validate currently returns nil and will be extended with real validation.
func (c *Config) Validate() error {
	return nil
}
