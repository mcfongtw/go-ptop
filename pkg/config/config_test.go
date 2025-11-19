//go:build unit

package config

import (
	"testing"
	"time"

	"go-ptop/pkg/formatter"
)

func Test_Default_ReturnsValidConfig(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name          string
		checkField    string
		expectedValue interface{}
	}{
		{"output_format", "OutputFormat", "tui"},
		{"update_interval", "UpdateInterval", 5 * time.Second},
		{"show_threads", "ShowThreads", true},
		{"show_mmaps", "ShowMMaps", true},
		{"top_n", "TopN", 50},
		{"sort_by", "SortBy", formatter.SortByRSS},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			cfg := Default()

			switch tc.checkField {
			case "OutputFormat":
				if cfg.OutputFormat != tc.expectedValue.(string) {
					t.Errorf("expected OutputFormat to be '%s', got %s", tc.expectedValue, cfg.OutputFormat)
				}
			case "UpdateInterval":
				if cfg.UpdateInterval != tc.expectedValue.(time.Duration) {
					t.Errorf("expected UpdateInterval to be %v, got %v", tc.expectedValue, cfg.UpdateInterval)
				}
			case "ShowThreads":
				if cfg.ShowThreads != tc.expectedValue.(bool) {
					t.Errorf("expected ShowThreads to be %v", tc.expectedValue)
				}
			case "ShowMMaps":
				if cfg.ShowMMaps != tc.expectedValue.(bool) {
					t.Errorf("expected ShowMMaps to be %v", tc.expectedValue)
				}
			case "TopN":
				if cfg.TopN != tc.expectedValue.(int) {
					t.Errorf("expected TopN to be %d, got %d", tc.expectedValue, cfg.TopN)
				}
			case "SortBy":
				if cfg.SortBy != tc.expectedValue.(formatter.SortColumn) {
					t.Errorf("expected SortBy to be %s, got %s", tc.expectedValue, cfg.SortBy)
				}
			}
		})
	}
}

func Test_Default_UnsetFields(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name       string
		checkField string
		expected   interface{}
	}{
		{"pid_unset", "PID", int32(0)},
		{"enable_nmt_false", "EnableNMT", false},
		{"enable_root_features_false", "EnableRootFeatures", false},
		{"verbose_false", "Verbose", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			cfg := Default()

			switch tc.checkField {
			case "PID":
				if cfg.PID != tc.expected.(int32) {
					t.Errorf("expected PID to be %d (unset), got %d", tc.expected, cfg.PID)
				}
			case "EnableNMT":
				if cfg.EnableNMT != tc.expected.(bool) {
					t.Errorf("expected EnableNMT to be %v by default", tc.expected)
				}
			case "EnableRootFeatures":
				if cfg.EnableRootFeatures != tc.expected.(bool) {
					t.Errorf("expected EnableRootFeatures to be %v by default", tc.expected)
				}
			case "Verbose":
				if cfg.Verbose != tc.expected.(bool) {
					t.Errorf("expected Verbose to be %v by default", tc.expected)
				}
			}
		})
	}
}

func Test_Validate_ValidConfigs(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name   string
		config Config
	}{
		{"default_config", Default()},
		{"empty_config", Config{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			err := tc.config.Validate()
			if err != nil {
				t.Errorf("expected nil error for %s, got: %v", tc.name, err)
			}
		})
	}
}

func Test_Config_FieldsCanBeSet(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name       string
		checkField string
		expected   interface{}
	}{
		{"pid", "PID", int32(1234)},
		{"enable_nmt", "EnableNMT", true},
		{"enable_root_features", "EnableRootFeatures", true},
		{"output_format", "OutputFormat", "json"},
		{"update_interval", "UpdateInterval", 10 * time.Second},
		{"show_threads", "ShowThreads", false},
		{"show_mmaps", "ShowMMaps", false},
		{"top_n", "TopN", 100},
		{"sort_by", "SortBy", formatter.SortByThreadID},
		{"verbose", "Verbose", true},
		{"log_level", "LogLevel", "debug"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

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

			switch tc.checkField {
			case "PID":
				if cfg.PID != tc.expected.(int32) {
					t.Errorf("expected PID %d, got %d", tc.expected, cfg.PID)
				}
			case "EnableNMT":
				if cfg.EnableNMT != tc.expected.(bool) {
					t.Errorf("expected EnableNMT to be %v", tc.expected)
				}
			case "EnableRootFeatures":
				if cfg.EnableRootFeatures != tc.expected.(bool) {
					t.Errorf("expected EnableRootFeatures to be %v", tc.expected)
				}
			case "OutputFormat":
				if cfg.OutputFormat != tc.expected.(string) {
					t.Errorf("expected OutputFormat '%s', got %s", tc.expected, cfg.OutputFormat)
				}
			case "UpdateInterval":
				if cfg.UpdateInterval != tc.expected.(time.Duration) {
					t.Errorf("expected UpdateInterval %v, got %v", tc.expected, cfg.UpdateInterval)
				}
			case "ShowThreads":
				if cfg.ShowThreads != tc.expected.(bool) {
					t.Errorf("expected ShowThreads to be %v", tc.expected)
				}
			case "ShowMMaps":
				if cfg.ShowMMaps != tc.expected.(bool) {
					t.Errorf("expected ShowMMaps to be %v", tc.expected)
				}
			case "TopN":
				if cfg.TopN != tc.expected.(int) {
					t.Errorf("expected TopN %d, got %d", tc.expected, cfg.TopN)
				}
			case "SortBy":
				if cfg.SortBy != tc.expected.(formatter.SortColumn) {
					t.Errorf("expected SortBy %s, got %s", tc.expected, cfg.SortBy)
				}
			case "Verbose":
				if cfg.Verbose != tc.expected.(bool) {
					t.Errorf("expected Verbose to be %v", tc.expected)
				}
			case "LogLevel":
				if cfg.LogLevel != tc.expected.(string) {
					t.Errorf("expected LogLevel '%s', got %s", tc.expected, cfg.LogLevel)
				}
			}
		})
	}
}
