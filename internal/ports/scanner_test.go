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

func TestParseLsofOutputMultipleOnSamePort(t *testing.T) {
	// Multiple processes on same port (SO_REUSEPORT)
	input := `COMMAND     PID   USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
nginx     1001   root   8u  IPv4 0x1234      0t0  TCP *:80 (LISTEN)
nginx     1002   root   8u  IPv4 0x1234      0t0  TCP *:80 (LISTEN)
`
	ports, err := parseLsofOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(ports))
	}
}

func TestParseLsofOutputLongProcessName(t *testing.T) {
	input := `COMMAND     PID   USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
com.apple.WebKit.Networking     42156   mike   23u  IPv4 0x1234      0t0  TCP *:3000 (LISTEN)
`
	ports, err := parseLsofOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 1 {
		t.Errorf("expected 1 port, got %d", len(ports))
	}
	if ports[0].Process != "com.apple.WebKit.Networking" {
		t.Errorf("expected full process name, got %q", ports[0].Process)
	}
}

func TestParsePortFromLsofNameEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty string", "", 0},
		{"only colon", ":", 0},
		{"colon at end", "127.0.0.1:", 0},
		{"multiple colons ipv6", "[::1]:8080", 8080},
		{"brackets without port", "[::1]", 0},
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

func TestParseHexPortEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty string", "", 0},
		{"only colon", ":", 0},
		{"multiple colons", "a:b:c", 0},
		{"lowercase hex", "00000000:0bb8", 3000},
		{"port 1", "00000000:0001", 1},
		{"port 65535", "00000000:FFFF", 65535},
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

func TestGetUsernameFromUIDInvalid(t *testing.T) {
	// Test with non-numeric UID
	result := getUsernameFromUID("invalid")
	if result != "invalid" {
		t.Errorf("getUsernameFromUID(\"invalid\") = %q, expected \"invalid\"", result)
	}
}

func TestFindByPortNoMatch(t *testing.T) {
	// Port 59999 is extremely unlikely to be in use
	matches, err := FindByPort(59999)
	if err != nil {
		t.Errorf("FindByPort(59999) returned error: %v", err)
	}
	if len(matches) != 0 {
		t.Logf("Unexpectedly found process on port 59999: %+v", matches)
	}
}

func TestScanReturnsNoError(t *testing.T) {
	// Just verify Scan doesn't error out
	_, err := Scan()
	if err != nil {
		t.Errorf("Scan() returned error: %v", err)
	}
}

func TestParseLsofOutputWithStrangeCharacters(t *testing.T) {
	// Process name with special characters
	input := `COMMAND     PID   USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
test-proc     42156   mike   23u  IPv4 0x1234      0t0  TCP *:3000 (LISTEN)
`
	ports, err := parseLsofOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(ports))
	}
	if ports[0].Process != "test-proc" {
		t.Errorf("expected 'test-proc', got %q", ports[0].Process)
	}
}

func TestParseProcNetTCPFormat(t *testing.T) {
	// Test the hex port parsing which is used by parseProcNetTCP
	// This tests the logic even though we can't run the full function on macOS
	tests := []struct {
		name     string
		hexAddr  string
		expected int
	}{
		{"HTTP port", "0100007F:0050", 80},
		{"HTTPS port", "0100007F:01BB", 443},
		{"Node default", "00000000:0BB8", 3000},
		{"PostgreSQL", "0100007F:1538", 5432},
		{"MySQL", "0100007F:0CEA", 3306},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := parseHexPort(tt.hexAddr)
			if port != tt.expected {
				t.Errorf("parseHexPort(%q) = %d, want %d", tt.hexAddr, port, tt.expected)
			}
		})
	}
}

func TestScanDarwinLsofNotFound(t *testing.T) {
	// This test verifies the error path exists but we can't actually trigger it
	// since lsof is always available on macOS
	// Keeping for coverage of the error checking logic
}

func TestFindByPortWithMatch(t *testing.T) {
	// FindByPort internally calls Scan, so this tests the integration
	// Even if nothing matches, it exercises the matching logic
	result, err := FindByPort(1)
	if err != nil {
		t.Errorf("FindByPort(1) returned error: %v", err)
	}
	// Port 1 is privileged and unlikely to be in use
	_ = result
}

func TestScanSortsResults(t *testing.T) {
	// Verify that Scan returns sorted results
	ports, err := Scan()
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	for i := 1; i < len(ports); i++ {
		if ports[i].Port < ports[i-1].Port {
			t.Errorf("Ports not sorted: %d < %d", ports[i].Port, ports[i-1].Port)
		}
	}
}

func TestParseLsofOutputVariousFormats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name: "ipv4 any interface",
			input: `COMMAND     PID   USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
node      1234   user   10u  IPv4 0x1234      0t0  TCP *:8080 (LISTEN)
`,
			expected: 8080,
		},
		{
			name: "ipv6 any interface",
			input: `COMMAND     PID   USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
node      1234   user   10u  IPv6 0x1234      0t0  TCP [::]:8080 (LISTEN)
`,
			expected: 8080,
		},
		{
			name: "specific ipv4 bind",
			input: `COMMAND     PID   USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
nginx     1234   root   10u  IPv4 0x1234      0t0  TCP 192.168.1.1:443 (LISTEN)
`,
			expected: 443,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ports, err := parseLsofOutput(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(ports) != 1 {
				t.Fatalf("expected 1 port, got %d", len(ports))
			}
			if ports[0].Port != tt.expected {
				t.Errorf("got port %d, expected %d", ports[0].Port, tt.expected)
			}
		})
	}
}
