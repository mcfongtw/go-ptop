//go:build unit

package config

import (
	"testing"
	"time"

	"go-ptop/pkg/formatter"
)

func TestDefaultReturnsValidConfig(t *testing.T) {
	cfg := Default()

	if cfg.OutputFormat != "tui" {
		t.Errorf("expected OutputFormat to be 'tui', got %s", cfg.OutputFormat)
	}
	if cfg.UpdateInterval != 5*time.Second {
		t.Errorf("expected UpdateInterval to be 5s, got %v", cfg.UpdateInterval)
	}
	if !cfg.ShowThreads {
		t.Error("expected ShowThreads to be true")
	}
	if !cfg.ShowMMaps {
		t.Error("expected ShowMMaps to be true")
	}
	if cfg.TopN != 50 {
		t.Errorf("expected TopN to be 50, got %d", cfg.TopN)
	}
	if cfg.SortBy != formatter.SortByRSS {
		t.Errorf("expected SortBy to be %s, got %s", formatter.SortByRSS, cfg.SortBy)
	}
}

func TestDefaultDoesNotSetPID(t *testing.T) {
	cfg := Default()

	if cfg.PID != 0 {
		t.Errorf("expected PID to be 0 (unset), got %d", cfg.PID)
	}
}

func TestDefaultDoesNotEnableOptionalFeatures(t *testing.T) {
	cfg := Default()

	if cfg.EnableNMT {
		t.Error("expected EnableNMT to be false by default")
	}
	if cfg.EnableRootFeatures {
		t.Error("expected EnableRootFeatures to be false by default")
	}
	if cfg.Verbose {
		t.Error("expected Verbose to be false by default")
	}
}

func TestValidateReturnsNilForDefault(t *testing.T) {
	cfg := Default()

	err := cfg.Validate()
	if err != nil {
		t.Errorf("expected nil error for default config, got: %v", err)
	}
}

func TestValidateReturnsNilForEmptyConfig(t *testing.T) {
	cfg := Config{}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("expected nil error for empty config, got: %v", err)
	}
}

func TestConfigFieldsCanBeSet(t *testing.T) {
	cfg := Config{
		PID:                1234,
		EnableNMT:          true,
		EnableRootFeatures: true,
		OutputFormat:       "json",
		UpdateInterval:     10 * time.Second,
		ShowThreads:        false,
		ShowMMaps:          false,
		TopN:               100,
		SortBy:             formatter.SortByThreadID,
		Verbose:            true,
		LogLevel:           "debug",
	}

	if cfg.PID != 1234 {
		t.Errorf("expected PID 1234, got %d", cfg.PID)
	}
	if !cfg.EnableNMT {
		t.Error("expected EnableNMT to be true")
	}
	if !cfg.EnableRootFeatures {
		t.Error("expected EnableRootFeatures to be true")
	}
	if cfg.OutputFormat != "json" {
		t.Errorf("expected OutputFormat 'json', got %s", cfg.OutputFormat)
	}
	if cfg.UpdateInterval != 10*time.Second {
		t.Errorf("expected UpdateInterval 10s, got %v", cfg.UpdateInterval)
	}
	if cfg.ShowThreads {
		t.Error("expected ShowThreads to be false")
	}
	if cfg.ShowMMaps {
		t.Error("expected ShowMMaps to be false")
	}
	if cfg.TopN != 100 {
		t.Errorf("expected TopN 100, got %d", cfg.TopN)
	}
	if cfg.SortBy != formatter.SortByThreadID {
		t.Errorf("expected SortBy %s, got %s", formatter.SortByThreadID, cfg.SortBy)
	}
	if !cfg.Verbose {
		t.Error("expected Verbose to be true")
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel 'debug', got %s", cfg.LogLevel)
	}
}
