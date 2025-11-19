//go:build linux

package jvm

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/shirou/gopsutil/process"
	"golang.org/x/sys/unix"
)

const attachSocketFormat = "/tmp/.java_pid%d"
const attachTriggerFormat = "/proc/%d/cwd/.attach_pid%d"

// linuxClient implements JVMClient via the HotSpot Attach socket on Linux.
type linuxClient struct {
	pid        int32
	socketPath string
}

// NewClient returns the Linux JVM attach client.
func NewClient() JVMClient {
	return &linuxClient{}
}

func (c *linuxClient) Connect(pid int32) error {
	c.pid = pid
	c.socketPath = fmt.Sprintf(attachSocketFormat, pid)

	exists, err := fileExists(c.socketPath)
	if err != nil {
		return err
	}
	if !exists {
		if err := startAttachServer(pid, c.socketPath); err != nil {
			return err
		}
	}
	return nil
}

func (c *linuxClient) Close() error {
	c.pid = 0
	c.socketPath = ""
	return nil
}

func (c *linuxClient) GetThreadDump() (*ThreadDump, error) {
	if c.pid == 0 {
		return nil, fmt.Errorf("client not connected")
	}

	output, err := c.executeAttachCommand("threaddump")
	if err != nil {
		return nil, err
	}

	threads, err := parseThreadDump(output)
	if err != nil {
		return nil, err
	}

	return &ThreadDump{
		Timestamp: time.Now(),
		Threads:   threads,
	}, nil
}

func (c *linuxClient) GetNMT(detail bool) (*NMTReport, error) {
	return nil, ErrUnsupported
}

func (c *linuxClient) GetHeapInfo() (*HeapInfo, error) {
	return nil, ErrUnsupported
}

func (c *linuxClient) ExecuteCommand(command string, args ...string) (string, error) {
	if c.pid == 0 {
		return "", fmt.Errorf("client not connected")
	}
	return c.executeAttachCommand(command, args...)
}

func (c *linuxClient) executeAttachCommand(command string, args ...string) (string, error) {
	addr := net.UnixAddr{Name: c.socketPath, Net: "unix"}
	conn, err := net.DialUnix("unix", nil, &addr)
	if err != nil {
		return "", fmt.Errorf("dial attach socket: %w", err)
	}
	defer conn.Close()

	if err := sendAttachString(conn, "1"); err != nil {
		return "", err
	}
	if err := sendAttachString(conn, command); err != nil {
		return "", err
	}

	// HotSpot expects exactly three argument slots.
	for i := 0; i < 3; i++ {
		var arg string
		if i < len(args) {
			arg = args[i]
		}
		if err := sendAttachString(conn, arg); err != nil {
			return "", err
		}
	}

	return readAttachString(conn)
}

func sendAttachString(conn net.Conn, value string) error {
	if !strings.HasSuffix(value, "\x00") {
		value += "\x00"
	}
	if _, err := conn.Write([]byte(value)); err != nil {
		return fmt.Errorf("write attach data: %w", err)
	}
	return nil
}

func readAttachString(conn net.Conn) (string, error) {
	var builder strings.Builder
	buf := make([]byte, 4096)

	for {
		n, err := conn.Read(buf)
		if n > 0 {
			builder.Write(buf[:n])
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("read attach data: %w", err)
		}
	}

	return builder.String(), nil
}

func startAttachServer(pid int32, socketPath string) error {
	trigger := fmt.Sprintf(attachTriggerFormat, pid, pid)
	if err := os.WriteFile(trigger, []byte{}, 0600); err != nil {
		return fmt.Errorf("create attach trigger: %w", err)
	}

	proc, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("locate pid %d: %w", pid, err)
	}
	if err := proc.SendSignal(unix.SIGQUIT); err != nil {
		return fmt.Errorf("send SIGQUIT: %w", err)
	}

	if err := waitForSocket(socketPath, time.Second, 60*time.Second); err != nil {
		return err
	}
	return nil
}

func waitForSocket(path string, pollInterval, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		exists, err := fileExists(path)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
		time.Sleep(pollInterval)
	}
	return fmt.Errorf("attach socket %s not created within timeout", path)
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
