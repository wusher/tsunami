# Usage

Complete CLI reference for Tsunami.

## Synopsis

```bash
tsunami [port...] [flags]
```

## Modes

| Mode | Command | Description |
|------|---------|-------------|
| Interactive | `tsunami` | TUI for browsing and killing |
| Kill | `tsunami 3000` | Kill process on port |
| List | `tsunami -l` | List all listening ports |

## Port Specification

```bash
# Single port
tsunami 3000

# Multiple ports
tsunami 3000 8080 5432

# Comma-separated
tsunami 3000,8080,5432

# Range (max 1000 ports)
tsunami 3000-3010

# Mixed
tsunami 3000 8080-8090 9000,9001
```

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--force` | `-f` | Skip confirmation prompt | `false` |
| `--list` | `-l` | List listening ports | `false` |
| `--quiet` | `-q` | Suppress output except errors | `false` |
| `--dry-run` | `-n` | Show what would be killed | `false` |
| `--signal` | `-s` | Signal to send | `TERM` |
| `--timeout` | `-t` | Escalation timeout | `2s` |
| `--all` | `-a` | Kill all processes on port | `false` |
| `--pid` | `-p` | Kill by PID instead of port | - |
| `--filter` | - | Filter by process/user (list mode) | - |
| `--json` | - | JSON output (list mode) | `false` |
| `--help` | `-h` | Show help | - |
| `--version` | `-v` | Show version | - |

## Signals

| Signal | Behavior |
|--------|----------|
| `TERM` | Graceful shutdown, escalates to KILL after timeout (default) |
| `KILL` | Immediate termination, cannot be caught |
| `INT` | Interrupt (like Ctrl+C) |
| `HUP` | Hangup signal |

```bash
# Default (SIGTERM with escalation)
tsunami 3000

# Immediate kill
tsunami 3000 -s KILL

# Custom escalation timeout
tsunami 3000 --timeout 5s
```

## Examples

### Basic Operations

```bash
# Interactive TUI
tsunami

# List all ports
tsunami -l

# Kill port 3000
tsunami 3000 -f

# Kill multiple ports
tsunami 3000 8080 5432 -f
```

### Filtering and Output

```bash
# Filter by process name
tsunami -l --filter node

# Filter by user
tsunami -l --filter user=postgres

# JSON output
tsunami -l --json
```

### Advanced

```bash
# Dry run
tsunami 3000-4000 --dry-run

# Kill all processes on port
tsunami 8080 --all -f

# Kill by PID
tsunami --pid 12345 -f

# Custom timeout
tsunami 5432 --timeout 10s -f

# Quiet mode for scripts
tsunami 3000 -f -q
```

## TUI Controls

| Key | Action |
|-----|--------|
| Type | Filter list |
| Up/Down | Navigate |
| Enter | Select to kill |
| Backspace | Delete filter char |
| Esc | Clear filter / Quit |

## Exit Codes

- `0` - Success
- `1` - Error

## Scripting

```bash
# Check exit code
if tsunami 3000 -f; then
  echo "Port freed"
fi

# Parse JSON output
tsunami -l --json | jq '.[].port'

# Kill all node processes
tsunami -l --filter node --json | jq -r '.[].port' | xargs -I {} tsunami {} -f
```

## Troubleshooting

**"No process listening on port X"**
- Use `tsunami -l` to see all listening ports

**"Permission denied"**
- Use `sudo tsunami 3000` for processes owned by other users

**Process won't die with SIGTERM**
- Try `tsunami 3000 -s KILL -f`
