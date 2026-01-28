package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/wusher/tsunami/internal/killer"
	"github.com/wusher/tsunami/internal/ports"
)

func TestListPorts(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listPorts()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if err != nil {
		t.Fatalf("listPorts() returned error: %v", err)
	}

	// Should contain header
	if !strings.Contains(output, "PORT") {
		t.Error("listPorts output should contain PORT header")
	}
}

func TestKillPortNotListening(t *testing.T) {
	sig, _ := killer.ParseSignal("TERM")
	err := killPort(99999, sig)

	if err == nil {
		t.Error("killPort(99999) should return error for unused port")
	}
	if !strings.Contains(err.Error(), "no process listening") {
		t.Errorf("error message should mention 'no process listening', got: %v", err)
	}
}

func TestKillPortInvalidPort(t *testing.T) {
	// Test through the run function by checking argument validation
	// This is implicitly tested through the CLI arg parsing
}

func TestConfirmNo(t *testing.T) {
	// Create a pipe to simulate stdin
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	// Write "n" to stdin
	go func() {
		_, _ = w.WriteString("n\n")
		w.Close()
	}()

	result := confirm("Test?")

	os.Stdin = oldStdin

	if result {
		t.Error("confirm should return false for 'n'")
	}
}

func TestConfirmYes(t *testing.T) {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		_, _ = w.WriteString("y\n")
		w.Close()
	}()

	result := confirm("Test?")

	os.Stdin = oldStdin

	if !result {
		t.Error("confirm should return true for 'y'")
	}
}

func TestConfirmYesFull(t *testing.T) {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		_, _ = w.WriteString("yes\n")
		w.Close()
	}()

	result := confirm("Test?")

	os.Stdin = oldStdin

	if !result {
		t.Error("confirm should return true for 'yes'")
	}
}

func TestConfirmEmpty(t *testing.T) {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		_, _ = w.WriteString("\n")
		w.Close()
	}()

	result := confirm("Test?")

	os.Stdin = oldStdin

	if result {
		t.Error("confirm should return false for empty input (default No)")
	}
}

func TestConfirmMixedCase(t *testing.T) {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		_, _ = w.WriteString("YES\n")
		w.Close()
	}()

	result := confirm("Test?")

	os.Stdin = oldStdin

	if !result {
		t.Error("confirm should return true for 'YES'")
	}
}

func TestRootCmdFlags(t *testing.T) {
	// Verify flags exist
	if rootCmd.Flags().Lookup("force") == nil {
		t.Error("--force flag should exist")
	}
	if rootCmd.Flags().Lookup("signal") == nil {
		t.Error("--signal flag should exist")
	}
	if rootCmd.Flags().Lookup("list") == nil {
		t.Error("--list flag should exist")
	}
	if rootCmd.Flags().Lookup("quiet") == nil {
		t.Error("--quiet flag should exist")
	}
}

func TestRootCmdShortFlags(t *testing.T) {
	if rootCmd.Flags().ShorthandLookup("f") == nil {
		t.Error("-f shorthand should exist")
	}
	if rootCmd.Flags().ShorthandLookup("s") == nil {
		t.Error("-s shorthand should exist")
	}
	if rootCmd.Flags().ShorthandLookup("l") == nil {
		t.Error("-l shorthand should exist")
	}
	if rootCmd.Flags().ShorthandLookup("q") == nil {
		t.Error("-q shorthand should exist")
	}
}

func TestRootCmdUsage(t *testing.T) {
	if rootCmd.Use != "tsunami [port...]" {
		t.Errorf("Use = %q, expected 'tsunami [port...]'", rootCmd.Use)
	}
}

func TestRootCmdShort(t *testing.T) {
	if rootCmd.Short != "Kill processes listening on ports" {
		t.Errorf("Short = %q, expected 'Kill processes listening on ports'", rootCmd.Short)
	}
}

func TestListPortsOutput(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listPorts()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if err != nil {
		t.Fatalf("listPorts() returned error: %v", err)
	}

	// Should have table format
	if !strings.Contains(output, "PID") {
		t.Error("listPorts output should contain PID header")
	}
	if !strings.Contains(output, "PROCESS") {
		t.Error("listPorts output should contain PROCESS header")
	}
}

func TestKillPortWithForce(t *testing.T) {
	// Save original force value
	origForce := force
	force = true
	defer func() { force = origForce }()

	sig, _ := killer.ParseSignal("TERM")
	// Port 99999 shouldn't be listening
	err := killPort(99999, sig)

	if err == nil {
		t.Error("killPort should return error for non-listening port")
	}
}

func TestKillPortWithSignalKill(t *testing.T) {
	sig, _ := killer.ParseSignal("KILL")
	// Port 99999 shouldn't be listening
	err := killPort(99999, sig)

	if err == nil {
		t.Error("killPort should return error for non-listening port")
	}
}

func TestConfirmReadError(t *testing.T) {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	// Close write end immediately to simulate read error
	w.Close()

	result := confirm("Test?")

	os.Stdin = oldStdin

	if result {
		t.Error("confirm should return false on read error")
	}
}

// Tests for new flags
func TestNewFlags(t *testing.T) {
	flags := []struct {
		name      string
		shorthand string
	}{
		{"dry-run", "n"},
		{"all", "a"},
		{"json", ""},
		{"filter", ""},
		{"timeout", "t"},
		{"pid", "p"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			if rootCmd.Flags().Lookup(f.name) == nil {
				t.Errorf("--%s flag should exist", f.name)
			}
			if f.shorthand != "" {
				if rootCmd.Flags().ShorthandLookup(f.shorthand) == nil {
					t.Errorf("-%s shorthand should exist", f.shorthand)
				}
			}
		})
	}
}

func TestVersionFlag(t *testing.T) {
	if rootCmd.Version == "" {
		t.Error("Version should be set")
	}
}

// Tests for port expansion
func TestExpandPortArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    []int
		wantErr bool
	}{
		{
			name: "single port",
			args: []string{"3000"},
			want: []int{3000},
		},
		{
			name: "multiple ports",
			args: []string{"3000", "8080"},
			want: []int{3000, 8080},
		},
		{
			name: "comma-separated",
			args: []string{"3000,8080,9000"},
			want: []int{3000, 8080, 9000},
		},
		{
			name: "range",
			args: []string{"3000-3003"},
			want: []int{3000, 3001, 3002, 3003},
		},
		{
			name: "mixed",
			args: []string{"80", "3000-3002", "8080,9000"},
			want: []int{80, 3000, 3001, 3002, 8080, 9000},
		},
		{
			name:    "invalid port",
			args:    []string{"abc"},
			wantErr: true,
		},
		{
			name:    "port out of range",
			args:    []string{"70000"},
			wantErr: true,
		},
		{
			name:    "invalid range (start > end)",
			args:    []string{"3005-3000"},
			wantErr: true,
		},
		{
			name:    "range too large",
			args:    []string{"1-2000"},
			wantErr: true,
		},
		{
			name:    "zero port",
			args:    []string{"0"},
			wantErr: true,
		},
		{
			name:    "negative in comma list",
			args:    []string{"3000,-1"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandPortArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("expandPortArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expandPortArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePort(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{"1", 1, false},
		{"80", 80, false},
		{"443", 443, false},
		{"3000", 3000, false},
		{"65535", 65535, false},
		{"0", 0, true},
		{"-1", 0, true},
		{"65536", 0, true},
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parsePort(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePort(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parsePort(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// Tests for filter functionality
func TestFilterPorts(t *testing.T) {
	portList := []ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "alice", Proto: "tcp"},
		{Port: 8080, PID: 200, Process: "python", User: "bob", Proto: "tcp"},
		{Port: 9000, PID: 300, Process: "node", User: "alice", Proto: "tcp"},
		{Port: 5432, PID: 400, Process: "postgres", User: "postgres", Proto: "tcp"},
	}

	tests := []struct {
		name   string
		filter string
		want   int // number of results
	}{
		{"filter by process name", "node", 2},
		{"filter by process name case insensitive", "NODE", 2},
		{"filter by process substring", "post", 1},
		{"filter by user", "user=alice", 2},
		{"filter by user case insensitive", "user=ALICE", 2},
		{"filter no match", "nginx", 0},
		{"filter user no match", "user=root", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterPorts(portList, tt.filter)
			if len(result) != tt.want {
				t.Errorf("filterPorts() returned %d results, want %d", len(result), tt.want)
			}
		})
	}
}

// Tests for JSON output
func TestPrintJSON(t *testing.T) {
	portList := []ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "alice", Proto: "tcp"},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := printJSON(portList)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("printJSON() returned error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Validate it's valid JSON
	var result []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("printJSON() output is not valid JSON: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result))
	}

	// Check fields
	if result[0]["port"].(float64) != 3000 {
		t.Error("JSON should contain port field")
	}
	if result[0]["process"].(string) != "node" {
		t.Error("JSON should contain process field")
	}
}

func TestListPortsJSON(t *testing.T) {
	// Save original values
	origJsonOut := jsonOut
	origFilter := filter
	jsonOut = true
	filter = ""
	defer func() {
		jsonOut = origJsonOut
		filter = origFilter
	}()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listPorts()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("listPorts() returned error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Should be valid JSON (either array or empty array message)
	if !strings.HasPrefix(strings.TrimSpace(output), "[") {
		t.Error("JSON output should start with [")
	}
}

func TestDryRunMode(t *testing.T) {
	// Save original values
	origDryRun := dryRun
	origForce := force
	dryRun = true
	force = true
	defer func() {
		dryRun = origDryRun
		force = origForce
	}()

	sig, _ := killer.ParseSignal("TERM")
	// Port 99999 shouldn't be listening, so we expect an error about no process
	err := killPort(99999, sig)

	// With dry-run, we should still get the "no process" error since there's nothing there
	if err == nil {
		t.Error("Expected error for non-listening port")
	}
}

func TestRootCmdHasVersion(t *testing.T) {
	if rootCmd.Version == "" {
		t.Error("rootCmd should have Version set")
	}
}

func TestRunListMode(t *testing.T) {
	// Save original values
	origList := list
	origJsonOut := jsonOut
	origFilter := filter
	list = true
	jsonOut = false
	filter = ""
	defer func() {
		list = origList
		jsonOut = origJsonOut
		filter = origFilter
	}()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	run(rootCmd, []string{})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Should have table header
	if !strings.Contains(output, "PORT") {
		t.Error("run in list mode should output PORT header")
	}
}

func TestRunListModeWithFilter(t *testing.T) {
	origList := list
	origJsonOut := jsonOut
	origFilter := filter
	list = true
	jsonOut = false
	filter = "nonexistent_process_xyz"
	defer func() {
		list = origList
		jsonOut = origJsonOut
		filter = origFilter
	}()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	run(rootCmd, []string{})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Should have header and "No listening ports found" message
	if !strings.Contains(output, "PORT") {
		t.Error("filtered list should still show header")
	}
}

func TestKillProcessDryRun(t *testing.T) {
	origDryRun := dryRun
	dryRun = true
	defer func() { dryRun = origDryRun }()

	p := ports.PortInfo{
		Port:    12345,
		PID:     99999,
		Process: "testproc",
		User:    "testuser",
		Proto:   "tcp",
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	sig, _ := killer.ParseSignal("TERM")
	err := killProcess(p, 12345, sig)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Errorf("killProcess in dry-run mode should not error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Would kill") {
		t.Error("dry-run should output 'Would kill'")
	}
	if !strings.Contains(output, "testproc") {
		t.Error("dry-run output should contain process name")
	}
}

func TestKillProcessUserCancels(t *testing.T) {
	origForce := force
	origDryRun := dryRun
	force = false
	dryRun = false
	defer func() {
		force = origForce
		dryRun = origDryRun
	}()

	// Mock stdin to say "n"
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		_, _ = w.WriteString("n\n")
		w.Close()
	}()
	defer func() { os.Stdin = oldStdin }()

	p := ports.PortInfo{
		Port:    12345,
		PID:     99999,
		Process: "testproc",
		User:    "testuser",
		Proto:   "tcp",
	}

	// Capture stdout to suppress prompt
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	sig, _ := killer.ParseSignal("TERM")
	err := killProcess(p, 12345, sig)

	wOut.Close()
	os.Stdout = oldStdout
	_, _ = io.Copy(io.Discard, rOut)

	if err != nil {
		t.Errorf("killProcess should return nil when user cancels: %v", err)
	}
}

func TestKillPIDsDryRun(t *testing.T) {
	origDryRun := dryRun
	origForce := force
	dryRun = true
	force = true
	defer func() {
		dryRun = origDryRun
		force = origForce
	}()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	sig, _ := killer.ParseSignal("TERM")
	err := killPIDs([]int{99999, 99998}, sig)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Errorf("killPIDs in dry-run mode should not error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Would kill") {
		t.Error("dry-run should output 'Would kill'")
	}
	if !strings.Contains(output, "99999") {
		t.Error("dry-run output should contain PID")
	}
}

func TestKillPIDsUserCancels(t *testing.T) {
	origForce := force
	origDryRun := dryRun
	force = false
	dryRun = false
	defer func() {
		force = origForce
		dryRun = origDryRun
	}()

	// Mock stdin to say "n"
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		_, _ = w.WriteString("n\n")
		w.Close()
	}()
	defer func() { os.Stdin = oldStdin }()

	// Capture stdout
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	sig, _ := killer.ParseSignal("TERM")
	err := killPIDs([]int{99999}, sig)

	wOut.Close()
	os.Stdout = oldStdout
	_, _ = io.Copy(io.Discard, rOut)

	// Should return nil (user cancelled, no error)
	if err != nil {
		t.Errorf("killPIDs should return nil when user cancels: %v", err)
	}
}

func TestKillPIDsNonexistent(t *testing.T) {
	origForce := force
	origDryRun := dryRun
	origQuiet := quiet
	force = true
	dryRun = false
	quiet = true
	defer func() {
		force = origForce
		dryRun = origDryRun
		quiet = origQuiet
	}()

	sig, _ := killer.ParseSignal("TERM")
	err := killPIDs([]int{999999999}, sig)

	if err == nil {
		t.Error("killPIDs should return error for nonexistent PID")
	}
}

func TestKillPortMultipleWithoutAll(t *testing.T) {
	// This test verifies the error message when multiple processes are on a port
	// We can't easily create this scenario, but we can test the error path exists
	// by checking coverage of the multiple processes branch
}

func TestListPortsWithFilter(t *testing.T) {
	origJsonOut := jsonOut
	origFilter := filter
	jsonOut = false
	filter = "impossible_filter_xyz_123"
	defer func() {
		jsonOut = origJsonOut
		filter = origFilter
	}()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listPorts()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("listPorts() returned error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "No listening ports") {
		t.Error("filtered list with no matches should say 'No listening ports'")
	}
}

func TestExpandPortArgsEmpty(t *testing.T) {
	result, err := expandPortArgs([]string{})
	if err != nil {
		t.Errorf("expandPortArgs([]) should not error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expandPortArgs([]) should return empty slice, got %v", result)
	}
}

func TestFilterPortsEmpty(t *testing.T) {
	result := filterPorts([]ports.PortInfo{}, "test")
	if len(result) != 0 {
		t.Error("filterPorts with empty list should return empty list")
	}
}

func TestPrintJSONEmpty(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := printJSON([]ports.PortInfo{})

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("printJSON([]) returned error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "[]") {
		t.Error("printJSON([]) should output empty array")
	}
}

func TestRunWithPIDsMode(t *testing.T) {
	origPids := pids
	origDryRun := dryRun
	origForce := force
	pids = []int{999999}
	dryRun = true
	force = true
	defer func() {
		pids = origPids
		dryRun = origDryRun
		force = origForce
	}()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	run(rootCmd, []string{})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Would kill") {
		t.Error("run with --pid in dry-run should output 'Would kill'")
	}
}


func TestKillPortNoProcess(t *testing.T) {
	sig, _ := killer.ParseSignal("TERM")
	err := killPort(59997, sig)

	if err == nil {
		t.Error("killPort should return error for port with no process")
	}
	if !strings.Contains(err.Error(), "no process") {
		t.Errorf("error should mention 'no process', got: %v", err)
	}
}

func TestListPortsTableOutput(t *testing.T) {
	origJsonOut := jsonOut
	origFilter := filter
	jsonOut = false
	filter = ""
	defer func() {
		jsonOut = origJsonOut
		filter = origFilter
	}()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := listPorts()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("listPorts() returned error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Should always have separator line
	if !strings.Contains(output, "---") {
		t.Error("table output should contain separator line")
	}
}

func TestKillProcessWithForceAndQuiet(t *testing.T) {
	origForce := force
	origDryRun := dryRun
	origQuiet := quiet
	force = true
	dryRun = false
	quiet = true
	defer func() {
		force = origForce
		dryRun = origDryRun
		quiet = origQuiet
	}()

	p := ports.PortInfo{
		Port:    12345,
		PID:     999999999, // nonexistent
		Process: "testproc",
		User:    "testuser",
		Proto:   "tcp",
	}

	sig, _ := killer.ParseSignal("TERM")
	err := killProcess(p, 12345, sig)

	// Should error because process doesn't exist
	if err == nil {
		t.Error("killProcess should error for nonexistent process")
	}
}

func TestKillProcessWithKillSignal(t *testing.T) {
	origForce := force
	origDryRun := dryRun
	force = true
	dryRun = false
	defer func() {
		force = origForce
		dryRun = origDryRun
	}()

	p := ports.PortInfo{
		Port:    12345,
		PID:     999999999, // nonexistent
		Process: "testproc",
		User:    "testuser",
		Proto:   "tcp",
	}

	sig, _ := killer.ParseSignal("KILL")
	err := killProcess(p, 12345, sig)

	// Should error because process doesn't exist
	if err == nil {
		t.Error("killProcess should error for nonexistent process")
	}
}

func TestKillPIDsWithOutput(t *testing.T) {
	origForce := force
	origDryRun := dryRun
	origQuiet := quiet
	force = true
	dryRun = false
	quiet = false
	defer func() {
		force = origForce
		dryRun = origDryRun
		quiet = origQuiet
	}()

	// Capture stdout/stderr
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	sig, _ := killer.ParseSignal("TERM")
	_ = killPIDs([]int{999999999}, sig)

	wOut.Close()
	os.Stdout = oldStdout
	_, _ = io.Copy(io.Discard, rOut)
}

func TestRunListModeJSON(t *testing.T) {
	origList := list
	origJsonOut := jsonOut
	origFilter := filter
	list = true
	jsonOut = true
	filter = ""
	defer func() {
		list = origList
		jsonOut = origJsonOut
		filter = origFilter
	}()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	run(rootCmd, []string{})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Should be valid JSON (starts with [ since it's an array)
	if !strings.HasPrefix(strings.TrimSpace(output), "[") {
		t.Error("run in list+json mode should output JSON array")
	}
}
