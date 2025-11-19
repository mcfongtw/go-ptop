//go:build !linux

package jvm

// NewClient returns a placeholder JVM client implementation for wiring.
func NewClient() JVMClient {
	return &noopClient{}
}

type noopClient struct{}

func (*noopClient) Connect(pid int32) error {
	return ErrUnsupported
}

func (*noopClient) Close() error {
	return nil
}

func (*noopClient) GetThreadDump() (*ThreadDump, error) {
	return nil, ErrUnsupported
}

func (*noopClient) GetNMT(detail bool) (*NMTReport, error) {
	return nil, ErrUnsupported
}

func (*noopClient) GetHeapInfo() (*HeapInfo, error) {
	return nil, ErrUnsupported
}

func (*noopClient) ExecuteCommand(command string, args ...string) (string, error) {
	return "", ErrUnsupported
}
