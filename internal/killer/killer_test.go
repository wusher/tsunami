package killer

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestParseSignal(t *testing.T) {
	tests := []struct {
		input    string
		expected Signal
		wantErr  bool
	}{
		// Valid signals - uppercase
		{"TERM", SIGTERM, false},
		{"KILL", SIGKILL, false},
		{"INT", SIGINT, false},
		{"HUP", SIGHUP, false},

		// Valid signals - lowercase
		{"term", SIGTERM, false},
		{"kill", SIGKILL, false},
		{"int", SIGINT, false},
		{"hup", SIGHUP, false},

		// Valid signals - mixed case
		{"Term", SIGTERM, false},
		{"Kill", SIGKILL, false},
		{"Int", SIGINT, false},
		{"Hup", SIGHUP, false},

		// With SIG prefix
		{"SIGTERM", SIGTERM, false},
		{"SIGKILL", SIGKILL, false},
		{"SIGINT", SIGINT, false},
		{"SIGHUP", SIGHUP, false},
		{"sigterm", SIGTERM, false},
		{"sighup", SIGHUP, false},

		// Invalid signals
		{"INVALID", "", true},
		{"USR1", "", true},
		{"", "", true},
		{"9", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseSignal(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseSignal(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseSignal(%q) unexpected error: %v", tt.input, err)
				return
			}

			if result != tt.expected {
				t.Errorf("ParseSignal(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSignalToSyscall(t *testing.T) {
	tests := []struct {
		signal   Signal
		expected syscall.Signal
	}{
		{SIGTERM, syscall.SIGTERM},
		{SIGKILL, syscall.SIGKILL},
		{SIGINT, syscall.SIGINT},
		{SIGHUP, syscall.SIGHUP},
		{Signal("UNKNOWN"), syscall.SIGTERM}, // Default
	}

	for _, tt := range tests {
		t.Run(string(tt.signal), func(t *testing.T) {
			result := tt.signal.toSyscall()
			if result != tt.expected {
				t.Errorf("%q.toSyscall() = %v, expected %v", tt.signal, result, tt.expected)
			}
		})
	}
}

func TestKillNonexistentProcess(t *testing.T) {
	// Use a PID that's very unlikely to exist
	err := Kill(999999999, SIGTERM)
	if err == nil {
		t.Error("Kill(999999999, SIGTERM) expected error for nonexistent process")
	}
}

func TestKillWithEscalation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process test in short mode")
	}

	// Start a process that ignores SIGTERM
	cmd := exec.Command("sh", "-c", "trap '' TERM; sleep 30")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start test process: %v", err)
	}

	pid := cmd.Process.Pid

	// Clean up in case of test failure
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	// Give the process time to set up signal handler
	time.Sleep(200 * time.Millisecond)

	// KillWithEscalation should TERM, wait, then KILL
	err := KillWithEscalation(pid)
	if err != nil {
		t.Errorf("KillWithEscalation(%d) returned error: %v", pid, err)
	}

	// Wait for process to fully exit
	_ = cmd.Wait()
}

func TestKillCooperativeProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process test in short mode")
	}

	// Start a process that exits on SIGTERM
	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start test process: %v", err)
	}

	pid := cmd.Process.Pid

	// Clean up in case of test failure
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	// Kill should succeed with SIGTERM
	err := Kill(pid, SIGTERM)
	if err != nil {
		t.Errorf("Kill(%d, SIGTERM) returned error: %v", pid, err)
	}

	// Wait for process to exit
	_ = cmd.Wait()
}

func TestIsProcessAlive(t *testing.T) {
	// Current process should be alive
	if !isProcessAlive(os.Getpid()) {
		t.Error("current process should be alive")
	}

	// Nonexistent process should not be alive
	if isProcessAlive(999999999) {
		t.Error("nonexistent process should not be alive")
	}
}

func TestKillWithEscalationNonexistent(t *testing.T) {
	err := KillWithEscalation(999999999)
	if err == nil {
		t.Error("KillWithEscalation(999999999) expected error for nonexistent process")
	}
}

func TestKillWithEscalationTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process test in short mode")
	}

	// Start a process that ignores SIGTERM
	cmd := exec.Command("sh", "-c", "trap '' TERM; sleep 30")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start test process: %v", err)
	}

	pid := cmd.Process.Pid

	// Clean up in case of test failure
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	// Give the process time to set up signal handler
	time.Sleep(200 * time.Millisecond)

	// KillWithEscalationTimeout with short timeout
	start := time.Now()
	err := KillWithEscalationTimeout(pid, 500*time.Millisecond)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("KillWithEscalationTimeout(%d) returned error: %v", pid, err)
	}

	// Should take at least 500ms (the timeout) but less than 3s
	if elapsed < 400*time.Millisecond {
		t.Errorf("KillWithEscalationTimeout completed too quickly: %v", elapsed)
	}
	if elapsed > 3*time.Second {
		t.Errorf("KillWithEscalationTimeout took too long: %v", elapsed)
	}

	// Wait for process to fully exit
	_ = cmd.Wait()
}

func TestKillWithEscalationTimeoutNonexistent(t *testing.T) {
	err := KillWithEscalationTimeout(999999999, time.Second)
	if err == nil {
		t.Error("KillWithEscalationTimeout(999999999) expected error for nonexistent process")
	}
}
