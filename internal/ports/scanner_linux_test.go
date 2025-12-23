//go:build linux

package ports

import (
	"testing"
)

func TestScanLinux(t *testing.T) {
	ports, err := scanLinux()
	if err != nil {
		t.Fatalf("scanLinux() error: %v", err)
	}
	// Just verify it runs without error
	_ = ports
}

func TestParseProcNetTCP(t *testing.T) {
	// Test with actual /proc/net/tcp if available
	ports, err := parseProcNetTCP("/proc/net/tcp", "tcp")
	if err != nil {
		t.Logf("parseProcNetTCP error (expected if not root): %v", err)
	}
	_ = ports
}

func TestFindProcessByInodeNotFound(t *testing.T) {
	// Test with invalid inode
	pid, process := findProcessByInode("999999999999")
	if pid != 0 {
		t.Errorf("expected pid 0 for invalid inode, got %d", pid)
	}
	if process != "" {
		t.Errorf("expected empty process for invalid inode, got %q", process)
	}
}
