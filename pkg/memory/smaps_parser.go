package memory

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func parseSMaps(r io.Reader) ([]ProcessMemorySegment, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var segments []ProcessMemorySegment
	var current *ProcessMemorySegment

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if isHeaderLine(line) {
			seg, err := parseSMapsHeader(line)
			if err != nil {
				return nil, err
			}
			segments = append(segments, seg)
			current = &segments[len(segments)-1]
			continue
		}

		if current == nil {
			continue
		}

		if err := parseSMapsStatLine(current, line); err != nil {
			return nil, err
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read smaps: %w", err)
	}

	return segments, nil
}

func isHeaderLine(line string) bool {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false
	}
	return strings.Contains(fields[0], "-")
}

func parseSMapsHeader(line string) (ProcessMemorySegment, error) {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return ProcessMemorySegment{}, fmt.Errorf("invalid smaps header: %s", line)
	}

	addr := strings.SplitN(fields[0], "-", 2)
	if len(addr) != 2 {
		return ProcessMemorySegment{}, fmt.Errorf("invalid address range: %s", fields[0])
	}

	start, err := strconv.ParseUint(addr[0], 16, 64)
	if err != nil {
		return ProcessMemorySegment{}, fmt.Errorf("parse start addr: %w", err)
	}

	end, err := strconv.ParseUint(addr[1], 16, 64)
	if err != nil {
		return ProcessMemorySegment{}, fmt.Errorf("parse end addr: %w", err)
	}

	offset, err := strconv.ParseUint(fields[2], 16, 64)
	if err != nil {
		return ProcessMemorySegment{}, fmt.Errorf("parse offset: %w", err)
	}

	inode, err := strconv.ParseUint(fields[4], 10, 64)
	if err != nil {
		return ProcessMemorySegment{}, fmt.Errorf("parse inode: %w", err)
	}

	path := ""
	if len(fields) > 5 {
		path = strings.Join(fields[5:], " ")
	}

	segment := ProcessMemorySegment{
		StartAddr:   start,
		EndAddr:     end,
		Permissions: fields[1],
		Offset:      offset,
		Device:      fields[3],
		Inode:       inode,
		Path:        path,
		SegmentType: classifySegment(path),
	}

	return segment, nil
}

func parseSMapsStatLine(seg *ProcessMemorySegment, line string) error {
	if strings.HasPrefix(line, "VmFlags") {
		return nil
	}

	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return nil
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	val, err := parseStatValue(value)
	if err != nil {
		return err
	}

	switch key {
	case "Size":
		seg.Size = val
	case "Rss":
		seg.RSS = val
	case "Pss":
		seg.PSS = val
	case "Shared_Clean":
		seg.SharedClean = val
	case "Shared_Dirty":
		seg.SharedDirty = val
	case "Private_Clean":
		seg.PrivateClean = val
	case "Private_Dirty":
		seg.PrivateDirty = val
	case "Referenced":
		seg.Referenced = val
	case "Anonymous":
		seg.Anonymous = val
	case "Swap":
		seg.Swap = val
	}

	return nil
}

func parseStatValue(field string) (uint64, error) {
	cleaned := strings.TrimSpace(strings.TrimSuffix(field, "kB"))
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		return 0, nil
	}
	val, err := strconv.ParseUint(cleaned, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse stat value '%s': %w", field, err)
	}
	return val, nil
}

func classifySegment(path string) SegmentType {
	switch {
	case strings.Contains(path, "[heap]"):
		return SegmentTypeHeap
	case strings.Contains(path, "[stack]"):
		return SegmentTypeStack
	case strings.Contains(path, "[vdso]"):
		return SegmentTypeVDSO
	case strings.Contains(path, "[vvar]"):
		return SegmentTypeVVar
	case strings.HasPrefix(path, "/"):
		return SegmentTypeMMap
	case path == "":
		return SegmentTypeAnon
	default:
		return SegmentTypeUnknown
	}
}
