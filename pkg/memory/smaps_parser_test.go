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
