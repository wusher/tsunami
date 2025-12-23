//go:build linux

package ports

import (
	"os"
	"path/filepath"
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

func TestParseProcNetTCPWithMockData(t *testing.T) {
	// Create a temporary file with sample /proc/net/tcp content
	content := `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 00000000:0BB8 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 12345 1 0000000000000000 100 0 0 10 0
   1: 0100007F:0050 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 67890 1 0000000000000000 100 0 0 10 0
   2: 00000000:1F90 00000000:0000 01 00000000:00000000 00:00000000 00000000  1000        0 11111 1 0000000000000000 100 0 0 10 0
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tcp")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	ports, err := parseProcNetTCP(tmpFile, "tcp")
	if err != nil {
		t.Fatalf("parseProcNetTCP() error: %v", err)
	}

	// State 0A = LISTEN, state 01 = ESTABLISHED
	// We should find sockets in LISTEN state, but findProcessByInode may not find matches
	// Just verify no error occurred
	_ = ports
}

func TestParseProcNetTCPMalformed(t *testing.T) {
	content := `  sl  local_address rem_address   st
   0: short
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tcp")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	ports, err := parseProcNetTCP(tmpFile, "tcp")
	if err != nil {
		t.Fatalf("parseProcNetTCP() error: %v", err)
	}

	// Should return empty list for malformed content
	if len(ports) != 0 {
		t.Errorf("expected 0 ports for malformed content, got %d", len(ports))
	}
}

func TestParseProcNetTCPFileNotFound(t *testing.T) {
	_, err := parseProcNetTCP("/nonexistent/path/tcp", "tcp")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestParseProcNetTCP6(t *testing.T) {
	// Test with /proc/net/tcp6 if available
	ports, err := parseProcNetTCP("/proc/net/tcp6", "tcp6")
	if err != nil {
		t.Logf("parseProcNetTCP6 error (expected if not available): %v", err)
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

func TestScanLinuxNoError(t *testing.T) {
	// Verify scanLinux handles both tcp and tcp6
	ports, err := scanLinux()
	if err != nil {
		t.Errorf("scanLinux() returned error: %v", err)
	}
	// Results may be empty if no ports are listening
	_ = ports
}
