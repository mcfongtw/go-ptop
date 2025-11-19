//go:build unit

package memory

import (
	"strings"
	"testing"
)

func Test_ParseSMaps_SingleSegment(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name          string
		input         string
		expectedCount int
		expectedType  SegmentType
		expectedRSS   uint64
		expectedSize  uint64
		expectedStart uint64
		expectedEnd   uint64
		expectedPerms string
	}{
		{
			name: "basic_mmap_segment",
			input: `00400000-00452000 r-xp 00000000 08:02 131073 /usr/bin/cat
Size:                 132 kB
Rss:                   12 kB
Pss:                    6 kB
Shared_Clean:          12 kB
Shared_Dirty:           0 kB
Private_Clean:          0 kB
Private_Dirty:          0 kB
Referenced:            12 kB
Anonymous:              0 kB
Swap:                   0 kB
`,
			expectedCount: 1,
			expectedType:  SegmentTypeMMap,
			expectedRSS:   12,
			expectedSize:  132,
			expectedStart: 0x00400000,
			expectedEnd:   0x00452000,
			expectedPerms: "r-xp",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			segments, err := parseSMaps(strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("parseSMaps returned error: %v", err)
			}

			if len(segments) != tc.expectedCount {
				t.Fatalf("expected %d segment(s), got %d", tc.expectedCount, len(segments))
			}

			seg := segments[0]
			if seg.StartAddr != tc.expectedStart || seg.EndAddr != tc.expectedEnd {
				t.Errorf("unexpected address range: %#x-%#x", seg.StartAddr, seg.EndAddr)
			}

			if seg.RSS != tc.expectedRSS || seg.Size != tc.expectedSize {
				t.Errorf("unexpected metrics rss=%d size=%d", seg.RSS, seg.Size)
			}

			if seg.SegmentType != tc.expectedType {
				t.Errorf("expected %s, got %s", tc.expectedType, seg.SegmentType)
			}

			if seg.Permissions != tc.expectedPerms {
				t.Errorf("expected permissions '%s', got '%s'", tc.expectedPerms, seg.Permissions)
			}
		})
	}
}

func Test_ParseSMaps_SegmentTypes(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name         string
		input        string
		expectedType SegmentType
	}{
		{
			name: "heap_segment",
			input: `01234000-01256000 rw-p 00000000 00:00 0 [heap]
Size:                 136 kB
Rss:                   48 kB
Pss:                   48 kB
Shared_Clean:           0 kB
Shared_Dirty:           0 kB
Private_Clean:          0 kB
Private_Dirty:         48 kB
Referenced:            48 kB
Anonymous:             48 kB
Swap:                   0 kB
`,
			expectedType: SegmentTypeHeap,
		},
		{
			name: "stack_segment",
			input: `7fff12345000-7fff12367000 rw-p 00000000 00:00 0 [stack]
Size:                 136 kB
Rss:                   24 kB
Pss:                   24 kB
Shared_Clean:           0 kB
Shared_Dirty:           0 kB
Private_Clean:          0 kB
Private_Dirty:         24 kB
Referenced:            24 kB
Anonymous:             24 kB
Swap:                   0 kB
`,
			expectedType: SegmentTypeStack,
		},
		{
			name: "vdso_segment",
			input: `7ffff7ffa000-7ffff7ffc000 r-xp 00000000 00:00 0 [vdso]
Size:                   8 kB
Rss:                    4 kB
Pss:                    0 kB
Shared_Clean:           4 kB
Shared_Dirty:           0 kB
Private_Clean:          0 kB
Private_Dirty:          0 kB
Referenced:             4 kB
Anonymous:              0 kB
Swap:                   0 kB
`,
			expectedType: SegmentTypeVDSO,
		},
		{
			name: "anonymous_segment",
			input: `7f1234000000-7f1234100000 rw-p 00000000 00:00 0
Size:                1024 kB
Rss:                  512 kB
Pss:                  512 kB
Shared_Clean:           0 kB
Shared_Dirty:           0 kB
Private_Clean:          0 kB
Private_Dirty:        512 kB
Referenced:           512 kB
Anonymous:            512 kB
Swap:                   0 kB
`,
			expectedType: SegmentTypeAnon,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			segments, err := parseSMaps(strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("parseSMaps returned error: %v", err)
			}

			if len(segments) != 1 {
				t.Fatalf("expected 1 segment, got %d", len(segments))
			}

			if segments[0].SegmentType != tc.expectedType {
				t.Errorf("expected %s, got %s", tc.expectedType, segments[0].SegmentType)
			}
		})
	}
}

func Test_ParseSMaps_MultipleSegments(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name          string
		input         string
		expectedCount int
		checkTypes    []SegmentType
	}{
		{
			name: "two_segments",
			input: `00400000-00452000 r-xp 00000000 08:02 131073 /usr/bin/cat
Size:                 132 kB
Rss:                   12 kB
Pss:                    6 kB
Shared_Clean:          12 kB
Shared_Dirty:           0 kB
Private_Clean:          0 kB
Private_Dirty:          0 kB
Referenced:            12 kB
Anonymous:              0 kB
Swap:                   0 kB
7fff12345000-7fff12367000 rw-p 00000000 00:00 0 [stack]
Size:                 136 kB
Rss:                   24 kB
Pss:                   24 kB
Shared_Clean:           0 kB
Shared_Dirty:           0 kB
Private_Clean:          0 kB
Private_Dirty:         24 kB
Referenced:            24 kB
Anonymous:             24 kB
Swap:                   0 kB
`,
			expectedCount: 2,
			checkTypes:    []SegmentType{SegmentTypeMMap, SegmentTypeStack},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			segments, err := parseSMaps(strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("parseSMaps returned error: %v", err)
			}

			if len(segments) != tc.expectedCount {
				t.Fatalf("expected %d segments, got %d", tc.expectedCount, len(segments))
			}

			for i, expectedType := range tc.checkTypes {
				if segments[i].SegmentType != expectedType {
					t.Errorf("segment %d: expected %s, got %s", i, expectedType, segments[i].SegmentType)
				}
			}
		})
	}
}

func Test_ParseSMaps_EmptyInput(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	t.Run("empty_string", func(t *testing.T) {
		teardownTest := setupTest(t)
		defer teardownTest(t)

		segments, err := parseSMaps(strings.NewReader(""))
		if err != nil {
			t.Fatalf("parseSMaps returned error for empty input: %v", err)
		}

		if len(segments) != 0 {
			t.Errorf("expected 0 segments for empty input, got %d", len(segments))
		}
	})
}

func Test_ParseSMaps_Permissions(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	tests := []struct {
		name          string
		input         string
		expectedPerms string
	}{
		{
			name: "rwxs_permissions",
			input: `00400000-00452000 rwxs 00000000 08:02 131073 /usr/bin/cat
Size:                 132 kB
Rss:                   12 kB
Pss:                    6 kB
Shared_Clean:          12 kB
Shared_Dirty:           0 kB
Private_Clean:          0 kB
Private_Dirty:          0 kB
Referenced:            12 kB
Anonymous:              0 kB
Swap:                   0 kB
`,
			expectedPerms: "rwxs",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			teardownTest := setupTest(t)
			defer teardownTest(t)

			segments, err := parseSMaps(strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("parseSMaps returned error: %v", err)
			}

			if segments[0].Permissions != tc.expectedPerms {
				t.Errorf("expected permissions '%s', got '%s'", tc.expectedPerms, segments[0].Permissions)
			}
		})
	}
}
