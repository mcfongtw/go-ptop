package proc

import (
	"fmt"
	"strconv"
	"strings"
)

// parseThreadStatData parses /proc stat content for a given TID.
func parseThreadStatData(tid int32, data string) (*ThreadStat, error) {
	fields := strings.Fields(data)
	if len(fields) < 30 {
		return nil, fmt.Errorf("unexpected stat format")
	}

	idx := 1
	for idx < len(fields) && !strings.HasSuffix(fields[idx], ")") {
		idx++
	}
	if idx >= len(fields) {
		return nil, fmt.Errorf("failed to locate command field")
	}

	nameFields := fields[1 : idx+1]
	name := strings.Join(nameFields, " ")
	name = strings.TrimPrefix(name, "(")
	name = strings.TrimSuffix(name, ")")

	if idx+26 >= len(fields) {
		return nil, fmt.Errorf("stat missing startstack field")
	}

	state := fields[idx+1]
	startStack, err := strconv.ParseUint(fields[idx+26], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse startstack: %w", err)
	}

	return &ThreadStat{
		TID:        tid,
		Name:       name,
		State:      state,
		StartStack: startStack,
	}, nil
}
