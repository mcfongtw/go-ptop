//go:build unit

package memory

import (
	"strings"
	"testing"
)

func TestParseSMapsParsesSingleSegment(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	const sample = `00400000-00452000 r-xp 00000000 08:02 131073 /usr/bin/cat
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
`

	segments, err := parseSMaps(strings.NewReader(sample))
	if err != nil {
		t.Fatalf("parseSMaps returned error: %v", err)
	}

	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}

	seg := segments[0]
	if seg.StartAddr != 0x00400000 || seg.EndAddr != 0x00452000 {
		t.Fatalf("unexpected address range: %#x-%#x", seg.StartAddr, seg.EndAddr)
	}

	if seg.RSS != 12 || seg.Size != 132 {
		t.Fatalf("unexpected metrics rss=%d size=%d", seg.RSS, seg.Size)
	}

	if seg.SegmentType != SegmentTypeMMap {
		t.Fatalf("expected SegmentTypeMMap, got %s", seg.SegmentType)
	}
}

func TestParseSMapsMultipleSegments(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	const sample = `00400000-00452000 r-xp 00000000 08:02 131073 /usr/bin/cat
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
`

	segments, err := parseSMaps(strings.NewReader(sample))
	if err != nil {
		t.Fatalf("parseSMaps returned error: %v", err)
	}

	if len(segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segments))
	}

	if segments[1].SegmentType != SegmentTypeStack {
		t.Errorf("expected SegmentTypeStack, got %s", segments[1].SegmentType)
	}
}

func TestParseSMapsHeapSegment(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	const sample = `01234000-01256000 rw-p 00000000 00:00 0 [heap]
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
`

	segments, err := parseSMaps(strings.NewReader(sample))
	if err != nil {
		t.Fatalf("parseSMaps returned error: %v", err)
	}

	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}

	if segments[0].SegmentType != SegmentTypeHeap {
		t.Errorf("expected SegmentTypeHeap, got %s", segments[0].SegmentType)
	}
}

func TestParseSMapsEmptyInput(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	segments, err := parseSMaps(strings.NewReader(""))
	if err != nil {
		t.Fatalf("parseSMaps returned error for empty input: %v", err)
	}

	if len(segments) != 0 {
		t.Errorf("expected 0 segments for empty input, got %d", len(segments))
	}
}

func TestParseSMapsVDSOSegment(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	const sample = `7ffff7ffa000-7ffff7ffc000 r-xp 00000000 00:00 0 [vdso]
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
`

	segments, err := parseSMaps(strings.NewReader(sample))
	if err != nil {
		t.Fatalf("parseSMaps returned error: %v", err)
	}

	if segments[0].SegmentType != SegmentTypeVDSO {
		t.Errorf("expected SegmentTypeVDSO, got %s", segments[0].SegmentType)
	}
}

func TestParseSMapsAnonymousSegment(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	const sample = `7f1234000000-7f1234100000 rw-p 00000000 00:00 0
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
`

	segments, err := parseSMaps(strings.NewReader(sample))
	if err != nil {
		t.Fatalf("parseSMaps returned error: %v", err)
	}

	if segments[0].SegmentType != SegmentTypeAnon {
		t.Errorf("expected SegmentTypeAnon, got %s", segments[0].SegmentType)
	}
}

func TestParseSMapsPermissions(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	const sample = `00400000-00452000 rwxs 00000000 08:02 131073 /usr/bin/cat
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
`

	segments, err := parseSMaps(strings.NewReader(sample))
	if err != nil {
		t.Fatalf("parseSMaps returned error: %v", err)
	}

	if segments[0].Permissions != "rwxs" {
		t.Errorf("expected permissions 'rwxs', got '%s'", segments[0].Permissions)
	}
}
