//go:build !linux

package memory

// NewMemoryReader returns a stub MemoryReader for unsupported platforms.
func NewMemoryReader() MemoryReader {
	return &noopMemoryReader{}
}

type noopMemoryReader struct{}

func (*noopMemoryReader) ReadSMaps(pid int32) ([]ProcessMemorySegment, error) {
	return nil, ErrUnsupported
}

func (*noopMemoryReader) ReadPageMap(pid int32, virtAddr, length uint64) ([]PageInfo, error) {
	return nil, ErrUnsupported
}

func (*noopMemoryReader) ReadNUMAMaps(pid int32) ([]NUMASegment, error) {
	return nil, ErrUnsupported
}
