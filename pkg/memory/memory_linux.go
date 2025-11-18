//go:build linux

package memory

import (
	"fmt"
	"os"
)

// linuxMemoryReader implements MemoryReader using /proc files.
type linuxMemoryReader struct{}

// NewMemoryReader returns the Linux implementation.
func NewMemoryReader() MemoryReader {
	return &linuxMemoryReader{}
}

func (r *linuxMemoryReader) ReadSMaps(pid int32) ([]ProcessMemorySegment, error) {
	path := fmt.Sprintf("/proc/%d/smaps", pid)
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open smaps: %w", err)
	}
	defer f.Close()

	segments, err := parseSMaps(f)
	if err != nil {
		return nil, fmt.Errorf("parse smaps: %w", err)
	}
	return segments, nil
}

func (r *linuxMemoryReader) ReadPageMap(pid int32, virtAddr, length uint64) ([]PageInfo, error) {
	return nil, ErrUnsupported
}

func (r *linuxMemoryReader) ReadNUMAMaps(pid int32) ([]NUMASegment, error) {
	return nil, ErrUnsupported
}
