//go:build unit

package proc

import "testing"

func TestParseThreadStatData(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	sample := "1234 (Reference Handler) R 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40 41 42"

	stat, err := parseThreadStatData(1234, sample)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if stat.Name != "Reference Handler" {
		t.Fatalf("unexpected name %s", stat.Name)
	}

	if stat.State != "R" {
		t.Fatalf("unexpected state %s", stat.State)
	}

	if stat.StartStack == 0 {
		t.Fatalf("expected non-zero startstack")
	}
}

func TestParseThreadStatDataDifferentStates(t *testing.T) {
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

func TestParseThreadStatDataThreadNameWithSpaces(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	sample := "9999 (GC Thread#5) S 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40 41 42"

	stat, err := parseThreadStatData(9999, sample)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if stat.Name != "GC Thread#5" {
		t.Errorf("expected name 'GC Thread#5', got '%s'", stat.Name)
	}
}

func TestParseThreadStatDataTIDCorrectlySet(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	sample := "4567 (main) R 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40 41 42"

	stat, err := parseThreadStatData(4567, sample)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if stat.TID != 4567 {
		t.Errorf("expected TID 4567, got %d", stat.TID)
	}
}

func TestParseThreadStatDataInvalidFormatTooShort(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	sample := "1234 (test)"

	_, err := parseThreadStatData(1234, sample)
	if err == nil {
		t.Fatal("expected error for invalid format, got nil")
	}
}

func TestParseThreadStatDataNoParentheses(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	sample := "1234 test R 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40 41 42"

	_, err := parseThreadStatData(1234, sample)
	if err == nil {
		t.Fatal("expected error for missing parentheses, got nil")
	}
}
