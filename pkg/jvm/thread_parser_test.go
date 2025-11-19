//go:build unit

package jvm

import "testing"

func Test_ParseThreadDump_ExtractsJavaThreads(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name           string
		input          string
		expectedCount  int
		checkFirstName string
		checkFirstOSID int32
		checkDaemon    bool
	}{
		{
			name: "two_daemon_threads",
			input: `"Reference Handler" #2 daemon prio=10 os_prio=0 tid=0x00007f2618009000 nid=0x695 waiting on condition [0x00007f25f3bfe000]
   java.lang.Thread.State: RUNNABLE

"Attach Listener" #7 daemon prio=9 os_prio=0 tid=0x00007f2614001000 nid=0x6b6 waiting on condition [0x0000000000000000]
   java.lang.Thread.State: RUNNABLE`,
			expectedCount:  2,
			checkFirstName: "Reference Handler",
			checkFirstOSID: 0x695,
			checkDaemon:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			threads, err := parseThreadDump(tc.input)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if len(threads) != tc.expectedCount {
				t.Fatalf("expected %d threads, got %d", tc.expectedCount, len(threads))
			}

			if threads[0].Name != tc.checkFirstName {
				t.Errorf("expected name '%s', got '%s'", tc.checkFirstName, threads[0].Name)
			}

			if threads[0].OSID != tc.checkFirstOSID {
				t.Errorf("expected OSID 0x%x, got 0x%x", tc.checkFirstOSID, threads[0].OSID)
			}

			if threads[1].Daemon != tc.checkDaemon {
				t.Errorf("expected daemon=%v, got %v", tc.checkDaemon, threads[1].Daemon)
			}

			if threads[1].JavaID == "" {
				t.Error("expected non-empty JavaID")
			}
		})
	}
}

func Test_ParseThreadDump_ErrorCases(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "empty_input",
			input:       "",
			expectError: true,
		},
		{
			name: "no_thread_lines",
			input: `Some random output
that doesn't contain
any thread information`,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			_, err := parseThreadDump(tc.input)
			if tc.expectError && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func Test_ParseThreadDump_NonDaemonThread(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name         string
		input        string
		expectDaemon bool
		expectName   string
		expectOSID   int32
	}{
		{
			name: "main_thread",
			input: `"main" #1 prio=5 os_prio=0 tid=0x00007f2618001000 nid=0x1234 runnable [0x00007f261c000000]
   java.lang.Thread.State: RUNNABLE`,
			expectDaemon: false,
			expectName:   "main",
			expectOSID:   0x1234,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			threads, err := parseThreadDump(tc.input)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if len(threads) != 1 {
				t.Fatalf("expected 1 thread, got %d", len(threads))
			}

			if threads[0].Daemon != tc.expectDaemon {
				t.Errorf("expected daemon=%v, got %v", tc.expectDaemon, threads[0].Daemon)
			}
			if threads[0].Name != tc.expectName {
				t.Errorf("expected name '%s', got '%s'", tc.expectName, threads[0].Name)
			}
			if threads[0].OSID != tc.expectOSID {
				t.Errorf("expected OSID 0x%x, got 0x%x", tc.expectOSID, threads[0].OSID)
			}
		})
	}
}

func Test_ParseThreadDump_ThreadState(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name        string
		input       string
		expectState string
	}{
		{
			name: "waiting_state",
			input: `"Waiting Thread" #5 daemon prio=5 os_prio=0 tid=0x00007f2618005000 nid=0x500 waiting on condition [0x00007f25f3000000]
   java.lang.Thread.State: WAITING (parking)`,
			expectState: "waiting",
		},
		{
			name: "runnable_state",
			input: `"Running Thread" #6 prio=5 os_prio=0 tid=0x00007f2618006000 nid=0x600 runnable [0x00007f25f3000000]
   java.lang.Thread.State: RUNNABLE`,
			expectState: "runnable",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			threads, err := parseThreadDump(tc.input)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if threads[0].ThreadState != tc.expectState {
				t.Errorf("expected state '%s', got '%s'", tc.expectState, threads[0].ThreadState)
			}
		})
	}
}

func Test_ParseThreadDump_Priorities(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name             string
		input            string
		expectedPriority int32
	}{
		{
			name: "high_priority",
			input: `"High Priority" #10 prio=10 os_prio=0 tid=0x00007f2618010000 nid=0xa00 runnable [0x00007f25f3000000]
   java.lang.Thread.State: RUNNABLE`,
			expectedPriority: 10,
		},
		{
			name: "low_priority",
			input: `"Low Priority" #11 prio=1 os_prio=0 tid=0x00007f2618011000 nid=0xb00 runnable [0x00007f25f3100000]
   java.lang.Thread.State: RUNNABLE`,
			expectedPriority: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			threads, err := parseThreadDump(tc.input)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if threads[0].Priority != tc.expectedPriority {
				t.Errorf("expected priority %d, got %d", tc.expectedPriority, threads[0].Priority)
			}
		})
	}
}
