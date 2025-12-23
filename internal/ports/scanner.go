// Package ports provides network port scanning functionality to discover
// processes listening on TCP ports. It supports both macOS (via lsof) and
// Linux (via /proc/net/tcp) platforms.
package ports

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// PortInfo represents a process listening on a port
type PortInfo struct {
	Port    int
	PID     int
	Process string
	User    string
	Proto   string // tcp, tcp6
}

// Scan returns all processes listening on TCP ports, sorted by port number
func Scan() ([]PortInfo, error) {
	var ports []PortInfo
	var err error

	switch runtime.GOOS {
	case "darwin":
		ports, err = scanDarwin()
	case "linux":
		ports, err = scanLinux()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if err != nil {
		return nil, err
	}

	// Sort by port number (low to high)
	sort.Slice(ports, func(i, j int) bool {
		return ports[i].Port < ports[j].Port
	})

	return ports, nil
}

// FindByPort returns all processes listening on a specific port
// Returns multiple results if SO_REUSEPORT is in use
func FindByPort(port int) ([]PortInfo, error) {
	all, err := Scan()
	if err != nil {
		return nil, err
	}

	var matches []PortInfo
	for _, p := range all {
		if p.Port == port {
			matches = append(matches, p)
		}
	}

	return matches, nil
}

// scanDarwin uses lsof to find listening ports on macOS
func scanDarwin() ([]PortInfo, error) {
	// Check if lsof is available
	if _, err := exec.LookPath("lsof"); err != nil {
		return nil, fmt.Errorf("lsof not found. Install with: brew install lsof")
	}

	// -iTCP: only TCP connections
	// -sTCP:LISTEN: only listening sockets
	// -n: no hostname resolution
	// -P: no port name resolution
	cmd := exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-n", "-P")
	output, err := cmd.Output()
	if err != nil {
		// lsof exits with 1 if no results, which is fine
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []PortInfo{}, nil
		}
		return nil, fmt.Errorf("lsof failed: %w", err)
	}

	return parseLsofOutput(string(output))
}

// parseLsofOutput parses lsof -iTCP -sTCP:LISTEN -n -P output
// Example line: node      42156  mike   23u  IPv4 0x1234  0t0  TCP *:3000 (LISTEN)
func parseLsofOutput(output string) ([]PortInfo, error) {
	var ports []PortInfo
	scanner := bufio.NewScanner(strings.NewReader(output))

	// Skip header line (COMMAND PID USER FD TYPE DEVICE SIZE/OFF NODE NAME)
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		process := fields[0]
		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		username := fields[2]

		// Parse the NAME field (last field before state)
		// Format: *:3000 or 127.0.0.1:3000 or [::1]:3000
		nameField := fields[8]
		port := parsePortFromLsofName(nameField)
		if port == 0 {
			continue
		}

		// Determine protocol from TYPE field
		proto := "tcp"
		if fields[4] == "IPv6" {
			proto = "tcp6"
		}

		ports = append(ports, PortInfo{
			Port:    port,
			PID:     pid,
			Process: process,
			User:    username,
			Proto:   proto,
		})
	}

	return ports, scanner.Err()
}

// parsePortFromLsofName extracts port from lsof NAME field
// Handles: *:3000, 127.0.0.1:3000, [::1]:3000
func parsePortFromLsofName(name string) int {
	// Find last colon
	idx := strings.LastIndex(name, ":")
	if idx == -1 {
		return 0
	}
	portStr := name[idx+1:]
	port, _ := strconv.Atoi(portStr)
	return port
}

// scanLinux parses /proc/net/tcp and /proc/net/tcp6
func scanLinux() ([]PortInfo, error) {
	var ports []PortInfo

	// Parse TCP (IPv4)
	tcp4, err := parseProcNetTCP("/proc/net/tcp", "tcp")
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	ports = append(ports, tcp4...)

	// Parse TCP6 (IPv6)
	tcp6, err := parseProcNetTCP("/proc/net/tcp6", "tcp6")
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	ports = append(ports, tcp6...)

	return ports, nil
}

// parseProcNetTCP parses /proc/net/tcp or /proc/net/tcp6
func parseProcNetTCP(path, proto string) ([]PortInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ports []PortInfo
	scanner := bufio.NewScanner(file)

	// Skip header line (sl local_address rem_address st tx_queue rx_queue tr tm->when retrnsmt uid timeout inode)
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		// Check if socket is in LISTEN state (0A)
		state := fields[3]
		if state != "0A" {
			continue
		}

		// Parse local address (format: hex_ip:hex_port)
		localAddr := fields[1]
		port := parseHexPort(localAddr)
		if port == 0 {
			continue
		}

		// Get inode
		inode := fields[9]

		// Find PID and process name from inode
		pid, process := findProcessByInode(inode)
		if pid == 0 {
			continue
		}

		// Get username from UID
		uid := fields[7]
		username := getUsernameFromUID(uid)

		ports = append(ports, PortInfo{
			Port:    port,
			PID:     pid,
			Process: process,
			User:    username,
			Proto:   proto,
		})
	}

	return ports, scanner.Err()
}

// parseHexPort extracts port from hex address format (ip:port)
func parseHexPort(addr string) int {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return 0
	}
	port, err := strconv.ParseInt(parts[1], 16, 32)
	if err != nil {
		return 0
	}
	return int(port)
}

// findProcessByInode searches /proc for the process using this socket inode
func findProcessByInode(inode string) (int, string) {
	target := fmt.Sprintf("socket:[%s]", inode)

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0, ""
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		fdPath := fmt.Sprintf("/proc/%d/fd", pid)
		fds, err := os.ReadDir(fdPath)
		if err != nil {
			continue
		}

		for _, fd := range fds {
			link, err := os.Readlink(fmt.Sprintf("%s/%s", fdPath, fd.Name()))
			if err != nil {
				continue
			}
			if link == target {
				// Found the process, get its name
				commPath := fmt.Sprintf("/proc/%d/comm", pid)
				comm, err := os.ReadFile(commPath)
				if err != nil {
					return pid, ""
				}
				return pid, strings.TrimSpace(string(comm))
			}
		}
	}

	return 0, ""
}

// getUsernameFromUID converts UID to username
func getUsernameFromUID(uid string) string {
	u, err := user.LookupId(uid)
	if err != nil {
		return uid
	}
	return u.Username
}
