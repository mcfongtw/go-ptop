//go:build !linux

package proc

// NewReader returns a placeholder ProcReader.
func NewReader() ProcReader {
	return &noopProcReader{}
}

type noopProcReader struct{}

func (*noopProcReader) Threads(pid int32) ([]KernelThread, error) {
	return nil, ErrUnsupported
}

func (*noopProcReader) ThreadStat(pid int32, tid int32) (*ThreadStat, error) {
	return nil, ErrUnsupported
}

func (*noopProcReader) ThreadIO(pid int32, tid int32) (*IOStats, error) {
	return nil, ErrUnsupported
}
