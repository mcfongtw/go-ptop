//go:build unit

package jvm

import "testing"

func TestParseThreadDumpExtractsJavaThreads(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	sample := `"Reference Handler" #2 daemon prio=10 os_prio=0 tid=0x00007f2618009000 nid=0x695 waiting on condition [0x00007f25f3bfe000]
   java.lang.Thread.State: RUNNABLE

"Attach Listener" #7 daemon prio=9 os_prio=0 tid=0x00007f2614001000 nid=0x6b6 waiting on condition [0x0000000000000000]
   java.lang.Thread.State: RUNNABLE`

	threads, err := parseThreadDump(sample)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(threads) != 2 {
		t.Fatalf("expected 2 threads, got %d", len(threads))
	}

	if threads[0].Name != "Reference Handler" || threads[0].OSID != 0x695 {
		t.Fatalf("unexpected first thread: %+v", threads[0])
	}

	if !threads[1].Daemon || threads[1].JavaID == "" {
		t.Fatalf("expected daemon thread with tid, got %+v", threads[1])
	}
}

func TestParseThreadDumpEmptyInput(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	_, err := parseThreadDump("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseThreadDumpNonDaemonThread(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	sample := `"main" #1 prio=5 os_prio=0 tid=0x00007f2618001000 nid=0x1234 runnable [0x00007f261c000000]
   java.lang.Thread.State: RUNNABLE`

	threads, err := parseThreadDump(sample)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(threads) != 1 {
		t.Fatalf("expected 1 thread, got %d", len(threads))
	}

	if threads[0].Daemon {
		t.Error("expected non-daemon thread")
	}
	if threads[0].Name != "main" {
		t.Errorf("expected name 'main', got '%s'", threads[0].Name)
	}
	if threads[0].OSID != 0x1234 {
		t.Errorf("expected OSID 0x1234, got 0x%x", threads[0].OSID)
	}
}

func TestParseThreadDumpExtractsThreadState(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	sample := `"Waiting Thread" #5 daemon prio=5 os_prio=0 tid=0x00007f2618005000 nid=0x500 waiting on condition [0x00007f25f3000000]
   java.lang.Thread.State: WAITING (parking)`

	threads, err := parseThreadDump(sample)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Parser extracts state from keywords in line, not from the State: line
	if threads[0].ThreadState != "waiting" {
		t.Errorf("expected state 'waiting', got '%s'", threads[0].ThreadState)
	}
}

func TestParseThreadDumpMultiplePriorities(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	sample := `"High Priority" #10 prio=10 os_prio=0 tid=0x00007f2618010000 nid=0xa00 runnable [0x00007f25f3000000]
   java.lang.Thread.State: RUNNABLE

"Low Priority" #11 prio=1 os_prio=0 tid=0x00007f2618011000 nid=0xb00 runnable [0x00007f25f3100000]
   java.lang.Thread.State: RUNNABLE`

	threads, err := parseThreadDump(sample)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(threads) != 2 {
		t.Fatalf("expected 2 threads, got %d", len(threads))
	}

	if threads[0].Priority != 10 {
		t.Errorf("expected priority 10, got %d", threads[0].Priority)
	}
	if threads[1].Priority != 1 {
		t.Errorf("expected priority 1, got %d", threads[1].Priority)
	}
}

func TestParseThreadDumpNoThreadLines(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	sample := `Some random output
that doesn't contain
any thread information`

	_, err := parseThreadDump(sample)
	if err == nil {
		t.Fatal("expected error when no threads found")
	}
}
