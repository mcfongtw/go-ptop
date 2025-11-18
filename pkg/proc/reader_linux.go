//go:build linux

package proc

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type linuxReader struct{}

// NewReader returns a Linux-backed ProcReader implementation.
func NewReader() ProcReader {
	return &linuxReader{}
}

func (r *linuxReader) Threads(pid int32) ([]KernelThread, error) {
	taskDir := fmt.Sprintf("/proc/%d/task", pid)
	entries, err := os.ReadDir(taskDir)
	if err != nil {
		return nil, fmt.Errorf("read task dir: %w", err)
	}

	threads := make([]KernelThread, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		tid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		stat, err := r.ThreadStat(pid, int32(tid))
		if err != nil {
			return nil, err
		}

		threads = append(threads, KernelThread{
			PID:        pid,
			TID:        int32(tid),
			Name:       stat.Name,
			State:      stat.State,
			StartStack: stat.StartStack,
		})
	}

	return threads, nil
}

func (r *linuxReader) ThreadStat(pid int32, tid int32) (*ThreadStat, error) {
	path := fmt.Sprintf("/proc/%d/task/%d/stat", pid, tid)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read stat: %w", err)
	}

	return parseThreadStatData(tid, string(data))
}

func (r *linuxReader) ThreadIO(pid int32, tid int32) (*IOStats, error) {
	path := fmt.Sprintf("/proc/%d/task/%d/io", pid, tid)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read io stats: %w", err)
	}
	defer file.Close()

	stats := &IOStats{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		value, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			continue
		}
		key := strings.TrimSuffix(parts[0], ":")
		switch key {
		case "syscr":
			stats.ReadCount = value
		case "syscw":
			stats.WriteCount = value
		case "read_bytes":
			stats.ReadBytes = value
		case "write_bytes":
			stats.WriteBytes = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan io stats: %w", err)
	}

	return stats, nil
}
