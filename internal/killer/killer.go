package killer

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
)

// Signal represents a process signal
type Signal string

const (
	SIGTERM Signal = "TERM"
	SIGKILL Signal = "KILL"
	SIGINT  Signal = "INT"
)

// ParseSignal parses a signal name (case-insensitive)
func ParseSignal(s string) (Signal, error) {
	switch strings.ToUpper(s) {
	case "TERM", "SIGTERM":
		return SIGTERM, nil
	case "KILL", "SIGKILL":
		return SIGKILL, nil
	case "INT", "SIGINT":
		return SIGINT, nil
	default:
		return "", fmt.Errorf("unknown signal: %s", s)
	}
}

// toSyscall converts Signal to syscall.Signal
func (s Signal) toSyscall() syscall.Signal {
	switch s {
	case SIGKILL:
		return syscall.SIGKILL
	case SIGINT:
		return syscall.SIGINT
	default:
		return syscall.SIGTERM
	}
}

// Kill sends a signal to a process
func Kill(pid int, sig Signal) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found: %w", err)
	}

	err = process.Signal(sig.toSyscall())
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied. Try sudo")
		}
		return fmt.Errorf("failed to kill process: %w", err)
	}

	return nil
}

// KillWithEscalation sends SIGTERM, waits, then SIGKILL if needed
func KillWithEscalation(pid int) error {
	// First, try SIGTERM
	err := Kill(pid, SIGTERM)
	if err != nil {
		return err
	}

	// Wait up to 2 seconds for process to exit
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if !isProcessAlive(pid) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Process still alive, send SIGKILL
	if isProcessAlive(pid) {
		return Kill(pid, SIGKILL)
	}

	return nil
}

// isProcessAlive checks if a process is still running
func isProcessAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Sending signal 0 checks if process exists without actually signaling
	err = process.Signal(syscall.Signal(0))
	return err == nil
}
