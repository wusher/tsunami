// Package main provides the CLI entry point for tsunami, a tool for killing
// processes bound to network ports. It supports both interactive TUI mode
// and direct command-line operation.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/wusher/tsunami/internal/killer"
	"github.com/wusher/tsunami/internal/ports"
	"github.com/wusher/tsunami/internal/tui"
)

// Version is set at build time via -ldflags
var Version = "dev"

var (
	force   bool
	signal  string
	list    bool
	quiet   bool
	dryRun  bool
	all     bool
	jsonOut bool
	filter  string
	timeout time.Duration
	pids    []int
)

var rootCmd = &cobra.Command{
	Use:     "tsunami [port...]",
	Short:   "Kill processes listening on ports",
	Version: Version,
	Long: `Tsunami is a CLI tool for killing processes bound to network ports.

When run without arguments, launches an interactive TUI for browsing and killing processes.
When given port arguments, kills processes on those ports directly.

By default, sends SIGTERM and escalates to SIGKILL after timeout if the process doesn't exit.

Examples:
  tsunami                    # Interactive TUI mode
  tsunami 3000               # Kill process on port 3000 (with confirmation)
  tsunami 3000 -f            # Kill without confirmation
  tsunami 3000 8080          # Kill processes on multiple ports
  tsunami 3000-3010          # Kill processes on ports 3000 through 3010
  tsunami 3000,8080,9000     # Comma-separated ports
  tsunami -l                 # List all listening ports
  tsunami -l --json          # List ports as JSON
  tsunami -l --filter node   # List only node processes
  tsunami 3000 -s KILL       # Send SIGKILL immediately
  tsunami 3000 --timeout 5s  # Wait 5s before escalating to SIGKILL
  tsunami --pid 1234         # Kill process by PID directly
  tsunami 3000 --dry-run     # Show what would be killed
  tsunami 3000 --all         # Kill all processes on port (when multiple)`,
	Args: cobra.ArbitraryArgs,
	Run:  run,
}

func init() {
	rootCmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	rootCmd.Flags().StringVarP(&signal, "signal", "s", "TERM", "Signal to send (TERM, KILL, INT, HUP)")
	rootCmd.Flags().BoolVarP(&list, "list", "l", false, "List listening ports and exit")
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Suppress output except errors")
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be killed without killing")
	rootCmd.Flags().BoolVarP(&all, "all", "a", false, "Kill all processes on port (when multiple)")
	rootCmd.Flags().BoolVar(&jsonOut, "json", false, "Output in JSON format (for --list)")
	rootCmd.Flags().StringVar(&filter, "filter", "", "Filter by process name, user, or user=<name> (for --list)")
	rootCmd.Flags().DurationVarP(&timeout, "timeout", "t", 2*time.Second, "Time to wait before escalating SIGTERM to SIGKILL")
	rootCmd.Flags().IntSliceVarP(&pids, "pid", "p", nil, "Kill processes by PID directly (can be repeated)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// run is the main command handler that dispatches to list, kill, or TUI mode
// based on the provided flags and arguments.
func run(cmd *cobra.Command, args []string) {
	// List mode
	if list {
		if err := listPorts(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Parse signal early to fail fast
	sig, err := killer.ParseSignal(signal)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// PID mode
	if len(pids) > 0 {
		if err := killPIDs(pids, sig); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// No args = interactive TUI mode
	if len(args) == 0 {
		if force {
			fmt.Fprintln(os.Stderr, "Error: --force requires port argument")
			os.Exit(1)
		}
		if dryRun {
			fmt.Fprintln(os.Stderr, "Error: --dry-run requires port argument")
			os.Exit(1)
		}
		if err := tui.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Expand port arguments (ranges and comma-separated)
	expandedPorts, err := expandPortArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Direct mode with port arguments
	var failures []string
	for _, port := range expandedPorts {
		if err := killPort(port, sig); err != nil {
			failures = append(failures, err.Error())
		}
	}

	if len(failures) > 0 {
		for _, f := range failures {
			fmt.Fprintf(os.Stderr, "Error: %s\n", f)
		}
		os.Exit(1)
	}
}

// expandPortArgs expands port arguments supporting ranges (3000-3005) and comma-separated (3000,8080,9000)
func expandPortArgs(args []string) ([]int, error) {
	var result []int
	rangePattern := regexp.MustCompile(`^(\d+)-(\d+)$`)

	for _, arg := range args {
		// Check for comma-separated values
		if strings.Contains(arg, ",") {
			parts := strings.Split(arg, ",")
			for _, part := range parts {
				port, err := parsePort(strings.TrimSpace(part))
				if err != nil {
					return nil, err
				}
				result = append(result, port)
			}
			continue
		}

		// Check for range
		if matches := rangePattern.FindStringSubmatch(arg); matches != nil {
			start, err := parsePort(matches[1])
			if err != nil {
				return nil, err
			}
			end, err := parsePort(matches[2])
			if err != nil {
				return nil, err
			}
			if start > end {
				return nil, fmt.Errorf("invalid port range: %s (start > end)", arg)
			}
			if end-start > 1000 {
				return nil, fmt.Errorf("port range too large: %s (max 1000 ports)", arg)
			}
			for p := start; p <= end; p++ {
				result = append(result, p)
			}
			continue
		}

		// Single port
		port, err := parsePort(arg)
		if err != nil {
			return nil, err
		}
		result = append(result, port)
	}

	return result, nil
}

// parsePort parses and validates a single port string
func parsePort(s string) (int, error) {
	port, err := strconv.Atoi(s)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("invalid port: %s (must be 1-65535)", s)
	}
	return port, nil
}

// listPorts displays all listening TCP ports in either table or JSON format.
// It respects the --filter and --json flags.
func listPorts() error {
	p, err := ports.Scan()
	if err != nil {
		return err
	}

	// Apply filter if specified
	if filter != "" {
		p = filterPorts(p, filter)
	}

	if len(p) == 0 {
		if jsonOut {
			fmt.Println("[]")
		} else {
			fmt.Println("No listening ports found")
		}
		return nil
	}

	if jsonOut {
		return printJSON(p)
	}

	fmt.Printf("%-8s %-10s %-20s %-15s %s\n", "PORT", "PID", "PROCESS", "USER", "PROTO")
	fmt.Println(strings.Repeat("-", 65))
	for _, port := range p {
		process := port.Process
		if len(process) > 20 {
			process = process[:17] + "..."
		}
		fmt.Printf("%-8d %-10d %-20s %-15s %s\n",
			port.Port, port.PID, process, port.User, port.Proto)
	}

	return nil
}

// filterPorts filters ports by process name or user
func filterPorts(portList []ports.PortInfo, f string) []ports.PortInfo {
	var result []ports.PortInfo

	// Check for user=<name> syntax
	if strings.HasPrefix(f, "user=") {
		userName := strings.TrimPrefix(f, "user=")
		for _, p := range portList {
			if strings.EqualFold(p.User, userName) {
				result = append(result, p)
			}
		}
		return result
	}

	// Default: filter by process name (case-insensitive substring match)
	fLower := strings.ToLower(f)
	for _, p := range portList {
		if strings.Contains(strings.ToLower(p.Process), fLower) {
			result = append(result, p)
		}
	}
	return result
}

// printJSON outputs port list as JSON
func printJSON(portList []ports.PortInfo) error {
	type jsonPort struct {
		Port    int    `json:"port"`
		PID     int    `json:"pid"`
		Process string `json:"process"`
		User    string `json:"user"`
		Proto   string `json:"proto"`
	}

	output := make([]jsonPort, len(portList))
	for i, p := range portList {
		output[i] = jsonPort{
			Port:    p.Port,
			PID:     p.PID,
			Process: p.Process,
			User:    p.User,
			Proto:   p.Proto,
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

// killPort finds and kills processes listening on the specified port.
// It handles confirmation prompts, dry-run mode, and multiple processes.
func killPort(port int, sig killer.Signal) error {
	matches, err := ports.FindByPort(port)
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		return fmt.Errorf("no process listening on port %d", port)
	}

	// Multiple processes on same port
	if len(matches) > 1 && !all {
		var pidList []string
		for _, m := range matches {
			pidList = append(pidList, strconv.Itoa(m.PID))
		}
		return fmt.Errorf("multiple processes on port %d: %s. Use --all to kill all",
			port, strings.Join(pidList, ", "))
	}

	// Kill all matching processes
	for _, p := range matches {
		if err := killProcess(p, port, sig); err != nil {
			return err
		}
	}

	return nil
}

// killProcess handles the actual killing of a single process
func killProcess(p ports.PortInfo, port int, sig killer.Signal) error {
	// Dry run mode
	if dryRun {
		fmt.Printf("Would kill: %s (PID %d) on port %d with signal %s\n", p.Process, p.PID, port, sig)
		return nil
	}

	// Confirmation
	if !force {
		if !confirm(fmt.Sprintf("Kill %s (PID %d) on port %d?", p.Process, p.PID, port)) {
			return nil // User cancelled
		}
	}

	// Kill the process
	var killErr error
	if sig == killer.SIGTERM {
		killErr = killer.KillWithEscalationTimeout(p.PID, timeout)
	} else {
		killErr = killer.Kill(p.PID, sig)
	}

	if killErr != nil {
		return killErr
	}

	if !quiet {
		fmt.Printf("Killed %s (PID %d) on port %d\n", p.Process, p.PID, port)
	}

	return nil
}

// killPIDs kills processes by their PIDs directly
func killPIDs(pidList []int, sig killer.Signal) error {
	var failures []string

	for _, pid := range pidList {
		if dryRun {
			fmt.Printf("Would kill: PID %d with signal %s\n", pid, sig)
			continue
		}

		if !force {
			if !confirm(fmt.Sprintf("Kill PID %d?", pid)) {
				continue // User cancelled
			}
		}

		var killErr error
		if sig == killer.SIGTERM {
			killErr = killer.KillWithEscalationTimeout(pid, timeout)
		} else {
			killErr = killer.Kill(pid, sig)
		}

		if killErr != nil {
			failures = append(failures, fmt.Sprintf("PID %d: %v", pid, killErr))
			continue
		}

		if !quiet {
			fmt.Printf("Killed PID %d\n", pid)
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("failed to kill: %s", strings.Join(failures, "; "))
	}
	return nil
}

// confirm prompts the user for confirmation and returns true if they respond
// with "y" or "yes" (case-insensitive). Default is "no" on empty input.
func confirm(msg string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N] ", msg)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
