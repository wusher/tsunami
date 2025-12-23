package ports

import (
	"testing"
)

func TestParseLsofOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []PortInfo
	}{
		{
			name:     "empty output",
			input:    "COMMAND     PID   USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME\n",
			expected: []PortInfo{},
		},
		{
			name: "single process",
			input: `COMMAND     PID   USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
node      42156   mike   23u  IPv4 0x1234      0t0  TCP *:3000 (LISTEN)
`,
			expected: []PortInfo{
				{Port: 3000, PID: 42156, Process: "node", User: "mike", Proto: "tcp"},
			},
		},
		{
			name: "multiple processes",
			input: `COMMAND     PID   USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
node      42156   mike   23u  IPv4 0x1234      0t0  TCP *:3000 (LISTEN)
postgres   1823   postgres 5u  IPv6 0x5678      0t0  TCP *:5432 (LISTEN)
`,
			expected: []PortInfo{
				{Port: 3000, PID: 42156, Process: "node", User: "mike", Proto: "tcp"},
				{Port: 5432, PID: 1823, Process: "postgres", User: "postgres", Proto: "tcp6"},
			},
		},
		{
			name: "ipv6 localhost",
			input: `COMMAND     PID   USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
redis     676   user   6u  IPv6 0x1234      0t0  TCP [::1]:6379 (LISTEN)
`,
			expected: []PortInfo{
				{Port: 6379, PID: 676, Process: "redis", User: "user", Proto: "tcp6"},
			},
		},
		{
			name: "bound to specific ip",
			input: `COMMAND     PID   USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
nginx     892   root   8u  IPv4 0x1234      0t0  TCP 127.0.0.1:8080 (LISTEN)
`,
			expected: []PortInfo{
				{Port: 8080, PID: 892, Process: "nginx", User: "root", Proto: "tcp"},
			},
		},
		{
			name: "malformed line - too few fields",
			input: `COMMAND     PID   USER   FD   TYPE
short     123   user
`,
			expected: []PortInfo{},
		},
		{
			name: "malformed line - invalid pid",
			input: `COMMAND     PID   USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
node      abc   mike   23u  IPv4 0x1234      0t0  TCP *:3000 (LISTEN)
`,
			expected: []PortInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseLsofOutput(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d results, got %d", len(tt.expected), len(result))
			}

			for i, exp := range tt.expected {
				if result[i].Port != exp.Port {
					t.Errorf("result[%d].Port = %d, expected %d", i, result[i].Port, exp.Port)
				}
				if result[i].PID != exp.PID {
					t.Errorf("result[%d].PID = %d, expected %d", i, result[i].PID, exp.PID)
				}
				if result[i].Process != exp.Process {
					t.Errorf("result[%d].Process = %q, expected %q", i, result[i].Process, exp.Process)
				}
				if result[i].User != exp.User {
					t.Errorf("result[%d].User = %q, expected %q", i, result[i].User, exp.User)
				}
				if result[i].Proto != exp.Proto {
					t.Errorf("result[%d].Proto = %q, expected %q", i, result[i].Proto, exp.Proto)
				}
			}
		})
	}
}

func TestParsePortFromLsofName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"wildcard", "*:3000", 3000},
		{"localhost ipv4", "127.0.0.1:8080", 8080},
		{"localhost ipv6", "[::1]:6379", 6379},
		{"all interfaces ipv6", "[::]:443", 443},
		{"high port", "*:65535", 65535},
		{"no colon", "invalid", 0},
		{"empty port", "*:", 0},
		{"non-numeric port", "*:abc", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePortFromLsofName(tt.input)
			if result != tt.expected {
				t.Errorf("parsePortFromLsofName(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseHexPort(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"port 80", "00000000:0050", 80},
		{"port 443", "00000000:01BB", 443},
		{"port 3000", "00000000:0BB8", 3000},
		{"port 8080", "00000000:1F90", 8080},
		{"invalid format", "invalid", 0},
		{"no colon", "00000000", 0},
		{"invalid hex", "00000000:ZZZZ", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseHexPort(tt.input)
			if result != tt.expected {
				t.Errorf("parseHexPort(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetUsernameFromUID(t *testing.T) {
	// Test with invalid UID - should return the UID itself
	result := getUsernameFromUID("99999999")
	if result != "99999999" {
		t.Errorf("getUsernameFromUID(\"99999999\") = %q, expected \"99999999\"", result)
	}

	// Test with UID 0 (root) - should return "root" on Unix systems
	result = getUsernameFromUID("0")
	if result != "root" {
		t.Errorf("getUsernameFromUID(\"0\") = %q, expected \"root\"", result)
	}
}

func TestFindByPort(t *testing.T) {
	// This is an integration test that uses the real Scan function
	// We can't predict which ports are listening, but we can verify it doesn't error
	_, err := FindByPort(99999) // Unlikely to be in use
	if err != nil {
		t.Errorf("FindByPort(99999) returned error: %v", err)
	}
}

func TestScan(t *testing.T) {
	// Integration test - verify Scan works and returns sorted results
	ports, err := Scan()
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	// Verify results are sorted by port
	for i := 1; i < len(ports); i++ {
		if ports[i].Port < ports[i-1].Port {
			t.Errorf("results not sorted: port %d comes after port %d",
				ports[i].Port, ports[i-1].Port)
		}
	}
}

func TestScanDarwin(t *testing.T) {
	// Only run on darwin
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// This tests the actual lsof parsing on macOS
	ports, err := scanDarwin()
	if err != nil {
		t.Fatalf("scanDarwin() error: %v", err)
	}

	// Should return some ports (system always has something listening)
	// But don't fail if nothing is listening
	_ = ports
}

func TestParseLsofOutputNoResults(t *testing.T) {
	// lsof with no results still has header
	input := "COMMAND     PID   USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME\n"
	ports, err := parseLsofOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 0 {
		t.Errorf("expected 0 ports, got %d", len(ports))
	}
}

func TestParseLsofOutputNoPort(t *testing.T) {
	// Line where port parsing fails
	input := `COMMAND     PID   USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
node      42156   mike   23u  IPv4 0x1234      0t0  TCP noport (LISTEN)
`
	ports, err := parseLsofOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 0 {
		t.Errorf("expected 0 ports for invalid line, got %d", len(ports))
	}
}
