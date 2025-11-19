//go:build unit

package proc

import "testing"

func Test_ParseThreadStatData_BasicParsing(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name          string
		tid           int32
		input         string
		expectedName  string
		expectedState string
		expectedTID   int32
	}{
		{
			name:          "reference_handler",
			tid:           1234,
			input:         "1234 (Reference Handler) R 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40 41 42",
			expectedName:  "Reference Handler",
			expectedState: "R",
			expectedTID:   1234,
		},
		{
			name:          "gc_thread_with_hash",
			tid:           9999,
			input:         "9999 (GC Thread#5) S 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40 41 42",
			expectedName:  "GC Thread#5",
			expectedState: "S",
			expectedTID:   9999,
		},
		{
			name:          "main_thread",
			tid:           4567,
			input:         "4567 (main) R 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40 41 42",
			expectedName:  "main",
			expectedState: "R",
			expectedTID:   4567,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			stat, err := parseThreadStatData(tc.tid, tc.input)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if stat.Name != tc.expectedName {
				t.Errorf("expected name '%s', got '%s'", tc.expectedName, stat.Name)
			}

			if stat.State != tc.expectedState {
				t.Errorf("expected state '%s', got '%s'", tc.expectedState, stat.State)
			}

			if stat.TID != tc.expectedTID {
				t.Errorf("expected TID %d, got %d", tc.expectedTID, stat.TID)
			}

			if stat.StartStack == 0 {
				t.Error("expected non-zero StartStack")
			}
		})
	}
}

func Test_ParseThreadStatData_DifferentStates(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name     string
		state    string
		expected string
	}{
		{"sleeping", "S", "S"},
		{"running", "R", "R"},
		{"disk_sleep", "D", "D"},
		{"stopped", "T", "T"},
		{"zombie", "Z", "Z"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			sample := "5678 (test) " + tc.state + " 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40 41 42"
			stat, err := parseThreadStatData(5678, sample)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if stat.State != tc.expected {
				t.Errorf("expected state %s, got %s", tc.expected, stat.State)
			}
		})
	}
}

func Test_ParseThreadStatData_ErrorCases(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name        string
		tid         int32
		input       string
		expectError bool
	}{
		{
			name:        "too_short_input",
			tid:         1234,
			input:       "1234 (test)",
			expectError: true,
		},
		{
			name:        "no_parentheses",
			tid:         1234,
			input:       "1234 test R 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40 41 42",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			_, err := parseThreadStatData(tc.tid, tc.input)
			if tc.expectError && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
