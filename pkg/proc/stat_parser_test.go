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
