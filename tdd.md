# Tsunami Technical Design Document

## Overview

Tsunami is a Go CLI tool for killing processes bound to network ports. Two modes of operation: interactive TUI for discovery/selection, or direct port targeting.

## Modes

### Interactive Mode (no arguments)

```
$ tsunami
```

Opens a TUI displaying all processes listening on ports. User selects a process, confirms, process is killed, app exits.

### Direct Mode (port argument)

```
$ tsunami 3000
```

Finds and kills the process listening on port 3000. Exits with status 0 on success, 1 on failure.

## Architecture

```
cmd/
  tsunami/
    main.go           # Entry point, mode detection
internal/
  ports/
    scanner.go        # Port scanning and process resolution
  killer/
    killer.go         # Process termination logic
  tui/
    tui.go            # Interactive interface (bubbletea)
    model.go          # TUI state management
```

## Core Components

### Port Scanner (`internal/ports/scanner.go`)

Responsible for enumerating listening ports and mapping them to processes.

**Approach:** Parse `/proc/net/tcp` and `/proc/net/tcp6` on Linux. Use `lsof` on macOS/BSD.

**Data Structure:**

```go
type PortInfo struct {
    Port    int
    PID     int
    Process string
    User    string
    Proto   string // tcp, tcp6
}

func Scan() ([]PortInfo, error)
func FindByPort(port int) (*PortInfo, error)
```

### Process Killer (`internal/killer/killer.go`)

Handles process termination with configurable signal.

```go
type Signal string

const (
    SIGTERM Signal = "TERM"
    SIGKILL Signal = "KILL"
    SIGINT  Signal = "INT"
)

func Kill(pid int, sig Signal) error
func KillWithEscalation(pid int) error // TERM -> wait -> KILL
```

**Default Strategy (no signal flag):**
1. Send SIGTERM
2. Wait 2 seconds
3. Send SIGKILL if still alive
4. Return error if process cannot be killed (permissions)

**Explicit Signal:**
When `--signal` is specified, send only that signal. No escalation. User takes responsibility.

### TUI (`internal/tui/`)

Built with [Bubbletea](https://github.com/charmbracelet/bubbletea) for the interactive interface.

**Features:**
- List view of all listening ports
- Filterable/searchable
- Arrow key navigation
- Enter to select
- Confirmation prompt before kill
- Color coded by port range (system/user/ephemeral)

**Visual Design:**

```
┌─────────────────────────────────────────────────────────────┐
│  🌊 TSUNAMI                                                 │
│  Kill processes listening on ports                          │
├─────────────────────────────────────────────────────────────┤
│  Filter: _                                                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│    PORT     PID      PROCESS              USER              │
│  ──────────────────────────────────────────────────────     │
│  ▸ 3000     42156    node                 mike         ◀    │
│    5432     1823     postgres             postgres          │
│    8080     51002    java                 mike              │
│    443      892      nginx                root              │
│    22       445      sshd                 root              │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│  ↑/↓ navigate  │  enter select  │  / filter  │  q quit     │
└─────────────────────────────────────────────────────────────┘
```

**Confirmation Modal:**

```
┌─────────────────────────────────────────────┐
│                                             │
│   ⚠️  Kill process?                          │
│                                             │
│   node (PID 42156) on port 3000             │
│                                             │
│        [ Yes ]          [ No ]              │
│                                             │
└─────────────────────────────────────────────┘
```

**Color Scheme (lipgloss):**

| Element | Color |
|---------|-------|
| Header/Title | Cyan (#00FFFF) |
| Selected row background | Dark blue (#1E3A5F) |
| Selected row text | White bold |
| System ports (< 1024) | Red (#FF6B6B) |
| User ports (1024 to 49151) | Green (#69FF94) |
| Ephemeral ports (> 49151) | Yellow (#FFFC58) |
| Filter input | Magenta (#FF79C6) |
| Keybinds footer | Dim gray (#6272A4) |
| Borders | Gray (#44475A) |

**Styling Constants:**

```go
var (
    titleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#00FFFF")).
        MarginBottom(1)

    selectedStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("#1E3A5F")).
        Foreground(lipgloss.Color("#FFFFFF")).
        Bold(true)

    systemPortStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FF6B6B"))

    userPortStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#69FF94"))

    borderStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("#44475A")).
        Padding(1, 2)
)
```

**Responsive Layout:**

TUI adapts to terminal size. Minimum 60x20. Uses `tea.WindowSizeMsg` to reflow content. Table columns auto adjust widths based on content. Process names truncate with ellipsis if too long.

**Model State:**

```go
type Model struct {
    ports     []ports.PortInfo
    cursor    int
    filter    string
    confirm   bool
    selected  *ports.PortInfo
    err       error
    width     int
    height    int
}
```

## CLI Interface

Using [Cobra](https://github.com/spf13/cobra) for argument parsing.

```
tsunami [port] [flags]
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip confirmation prompt |
| `--signal` | `-s` | Signal to send (TERM, KILL, INT). Default: TERM |
| `--list` | `-l` | List listening ports and exit (no kill) |
| `--quiet` | `-q` | Suppress output except errors |

**Usage Examples:**

```bash
# Interactive TUI
tsunami

# Kill process on port 3000, with confirmation
tsunami 3000

# Kill without confirmation (scripting)
tsunami 3000 -f

# Kill multiple ports
tsunami 3000 8080 5432 -f

# Just list what's listening
tsunami -l

# Send SIGKILL immediately (skip SIGTERM)
tsunami 3000 -s KILL

# Quiet mode for scripts
tsunami 3000 -fq
```

**Argument Handling:**

```go
var rootCmd = &cobra.Command{
    Use:   "tsunami [port...]",
    Short: "Kill processes listening on ports",
    Args:  cobra.ArbitraryArgs,
    Run:   run,
}

func init() {
    rootCmd.Flags().BoolP("force", "f", false, "Skip confirmation")
    rootCmd.Flags().StringP("signal", "s", "TERM", "Signal to send")
    rootCmd.Flags().BoolP("list", "l", false, "List only, no kill")
    rootCmd.Flags().BoolP("quiet", "q", false, "Suppress output")
}
```

**Mode Logic:**

```
if --list flag:
    print port table, exit

if no args:
    if --force:
        error "force requires port argument"
    launch TUI

if args (ports):
    for each port:
        find process
        if --force or confirm():
            kill with --signal
```

## Error Handling

| Scenario | Exit Code | Message |
|----------|-----------|---------|
| Success | 0 | (none or "Killed PID X on port Y") |
| Port not in use | 1 | "No process listening on port X" |
| Permission denied | 1 | "Permission denied. Try sudo." |
| Invalid port | 1 | "Invalid port: X" |
| Invalid signal | 1 | "Unknown signal: X" |
| Force without port | 1 | "Force mode requires port argument" |
| User cancelled | 0 | (none) |
| Partial failure (multiple ports) | 1 | Lists failures, exits 1 |

## Dependencies

- `github.com/spf13/cobra` for CLI
- `github.com/charmbracelet/bubbletea` for TUI
- `github.com/charmbracelet/lipgloss` for styling
- `github.com/charmbracelet/bubbles` for list component

## Platform Support

| Platform | Method |
|----------|--------|
| Linux | `/proc/net/tcp*` parsing |
| macOS | `lsof -iTCP -sTCP:LISTEN -n -P` |
| Windows | Not supported (v1) |

## Security Considerations

- Requires elevated privileges to kill processes owned by other users
- No network access required
- No persistent storage
- Reads only from `/proc` or calls `lsof`

## Build and Distribution

```
go build -o tsunami cmd/tsunami/main.go
```

Single binary, no runtime dependencies. Cross compile for Linux/macOS with standard Go toolchain.

## Future Considerations

- UDP port support
- Process name matching (`tsunami --name node`)
- Port range support (`tsunami 3000:3010`)
- JSON output for scripting (`tsunami -l --json`)
- Windows support via netstat parsing
- Config file for default flags
- Shell completions (bash, zsh, fish)
