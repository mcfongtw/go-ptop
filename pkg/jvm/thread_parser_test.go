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
