package jvm

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var threadHeaderPattern = regexp.MustCompile(`"(?P<name>[^"]+)".*?prio=(?P<prio>\d+).*?tid=(?P<tid>0x[0-9a-f]+).*?nid=(?P<nid>0x[0-9a-f]+).*`)

func parseThreadDump(raw string) ([]JavaThread, error) {
	lines := strings.Split(raw, "\n")
	threads := make([]JavaThread, 0)

	for _, line := range lines {
		thread, ok := parseThreadHeader(line)
		if !ok {
			continue
		}
		threads = append(threads, thread)
	}

	if len(threads) == 0 {
		return nil, fmt.Errorf("no thread headers found in dump")
	}
	return threads, nil
}

func parseThreadHeader(line string) (JavaThread, bool) {
	matches := threadHeaderPattern.FindStringSubmatch(line)
	if matches == nil {
		return JavaThread{}, false
	}

	idx := make(map[string]int)
	for i, name := range threadHeaderPattern.SubexpNames() {
		idx[name] = i
	}

	thread := JavaThread{
		Name:   matches[idx["name"]],
		JavaID: matches[idx["tid"]],
	}

	if v, err := strconv.Atoi(matches[idx["prio"]]); err == nil {
		thread.Priority = int32(v)
	}

	if nid, err := strconv.ParseInt(matches[idx["nid"]], 0, 32); err == nil {
		thread.OSID = int32(nid)
	}

	if strings.Contains(line, "daemon") {
		thread.Daemon = true
	}
	if strings.Contains(line, "runnable") {
		thread.ThreadState = "runnable"
	} else if strings.Contains(line, "waiting") {
		thread.ThreadState = "waiting"
	}

	return thread, true
}
