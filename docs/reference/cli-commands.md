# CLI Command Reference

Complete reference for all Tsunami command-line options and arguments.

## Synopsis

```bash
tsunami [port...] [flags]
```

## Modes

Tsunami operates in three distinct modes based on the arguments and flags provided:

### Interactive TUI Mode

```bash
tsunami
```

Launches an interactive terminal user interface for browsing and killing processes.

**Requirements:**
- No port arguments
- No `--list` flag

### List Mode

```bash
tsunami --list [flags]
```

Lists all listening TCP ports without killing anything.

**Aliases:** `-l`, `--list`

### Direct Kill Mode

```bash
tsunami <port> [port...] [flags]
```

Directly kills processes on specified port(s).

**Requires:** One or more port arguments

## Arguments

### Port Specification

#### Single Port

```bash
tsunami 3000
```

Kills the process listening on port 3000.

#### Multiple Ports (Space-separated)

```bash
tsunami 3000 8080 5432
```

Kills processes on ports 3000, 8080, and 5432.

#### Multiple Ports (Comma-separated)

```bash
tsunami 3000,8080,5432
```

Alternative syntax using commas.

#### Port Range

```bash
tsunami 3000-3010
```

Kills processes on all ports from 3000 to 3010 (inclusive).

**Limitations:**
- Maximum range size: 1000 ports
- Start must be less than or equal to end

#### Mixed Syntax

```bash
tsunami 3000 8080-8090 9000,9001,9002
```

Combine different port specification methods.

### Port Validation

- Must be between 1 and 65535
- Must be valid integers
- Invalid ports cause immediate error

## Flags

### Core Flags

#### `--force`, `-f`

Skip confirmation prompts.

```bash
tsunami 3000 -f
```

**Default:** `false`

**Use cases:**
- Automation scripts
- CI/CD pipelines
- When you're certain about the target

:::warning
Use carefully - no confirmation means no safety net!
:::

#### `--list`, `-l`

List all listening ports and exit (no killing).

```bash
tsunami -l
```

**Default:** `false`

**Output:** Table format by default, JSON with `--json`

#### `--quiet`, `-q`

Suppress success output; only show errors.

```bash
tsunami 3000 -f -q
```

**Default:** `false`

**Useful for:** Scripts where you only care about failures

#### `--help`, `-h`

Show help message and exit.

```bash
tsunami --help
```

#### `--version`, `-v`

Show version and exit.

```bash
tsunami --version
```

### Signal Flags

#### `--signal <SIGNAL>`, `-s <SIGNAL>`

Specify which signal to send to the process.

```bash
tsunami 3000 -s KILL
```

**Default:** `TERM`

**Available signals:**
- `TERM` - Graceful termination (default)
- `KILL` - Forceful termination (cannot be caught)
- `INT` - Interrupt (like Ctrl+C)
- `HUP` - Hangup (often used for reload)

**Signal behavior:**
- `TERM`: Escalates to `KILL` after timeout if process doesn't exit
- Other signals: Sent once, no escalation

See [Signal Types](./signals.md) for details.

#### `--timeout <DURATION>`, `-t <DURATION>`

Time to wait before escalating SIGTERM to SIGKILL.

```bash
tsunami 3000 --timeout 5s
```

**Default:** `2s`

**Format:** Go duration format (e.g., `1s`, `500ms`, `2m`)

**Only applies to:** `SIGTERM` signal

**Examples:**
```bash
tsunami 3000 --timeout 1s     # 1 second
tsunami 3000 --timeout 500ms  # 500 milliseconds
tsunami 3000 --timeout 10s    # 10 seconds
```

### Process Selection Flags

#### `--all`, `-a`

Kill all processes on the specified port (when multiple exist).

```bash
tsunami 8080 --all
```

**Default:** `false`

**Behavior without flag:** Error when multiple processes detected

**Use case:** Port shared by multiple processes (e.g., load-balanced services)

#### `--pid <PID>`, `-p <PID>`

Kill process by PID instead of port.

```bash
# Single PID
tsunami --pid 12345

# Multiple PIDs
tsunami --pid 12345 --pid 12346
```

**Can be repeated** for multiple PIDs

**Incompatible with:** Port arguments

### Filter & Output Flags

#### `--filter <PATTERN>`

Filter ports by process name or user (list mode only).

```bash
# By process name (substring match)
tsunami -l --filter node

# By user (exact match)
tsunami -l --filter user=postgres
```

**Syntax:**
- `<name>`: Filter by process name (case-insensitive substring)
- `user=<name>`: Filter by username (case-insensitive exact match)

**Only applies to:** `--list` mode

#### `--json`

Output in JSON format (list mode only).

```bash
tsunami -l --json
```

**Only applies to:** `--list` mode

**Output format:**
```json
[
  {
    "port": 3000,
    "pid": 12345,
    "process": "node",
    "user": "youruser",
    "proto": "tcp"
  }
]
```

### Dry Run

#### `--dry-run`, `-n`

Show what would be killed without actually killing.

```bash
tsunami 3000 --dry-run
```

**Default:** `false`

**Output:**
```
Would kill: node (PID 12345) on port 3000 with signal TERM
```

**Use cases:**
- Testing ranges before execution
- Verifying targets
- Learning what would happen

## Flag Combinations

### Valid Combinations

```bash
# Force + quiet (common in scripts)
tsunami 3000 -f -q

# List + filter + JSON
tsunami -l --filter node --json

# Dry run + custom signal
tsunami 3000 --dry-run -s KILL

# Multiple ports + force + timeout
tsunami 3000 8080 -f --timeout 5s

# Kill all + force
tsunami 8080 --all -f
```

### Invalid Combinations

```bash
# ❌ Force without arguments
tsunami -f
# Error: --force requires port argument

# ❌ JSON without list
tsunami 3000 --json
# Error: --json requires --list

# ❌ Filter without list
tsunami --filter node
# Error: --filter requires --list

# ❌ Dry run without arguments
tsunami --dry-run
# Error: --dry-run requires port argument

# ❌ PID and port together
tsunami 3000 --pid 12345
# Error: cannot specify both port and PID
```

## Output Formats

### Table Format (Default)

```bash
$ tsunami -l
PORT     PID        PROCESS              USER            PROTO
-----------------------------------------------------------------
3000     12345      node                 youruser        tcp
8080     12346      python3              youruser        tcp
5432     12347      postgres             postgres        tcp
```

### JSON Format

```bash
$ tsunami -l --json
[
  {
    "port": 3000,
    "pid": 12345,
    "process": "node",
    "user": "youruser",
    "proto": "tcp"
  },
  {
    "port": 8080,
    "pid": 12346,
    "process": "python3",
    "user": "youruser",
    "proto": "tcp"
  }
]
```

### Success Output

```bash
$ tsunami 3000 -f
Killed node (PID 12345) on port 3000
```

### Quiet Mode (Success)

```bash
$ tsunami 3000 -f -q
# No output
```

### Error Output

Always written to `stderr`, even in quiet mode:

```bash
$ tsunami 3000
Error: no process listening on port 3000

$ tsunami 99999
Error: invalid port: 99999 (must be 1-65535)

$ tsunami 8080
Error: multiple processes on port 8080: 12345, 12346. Use --all to kill all
```

## Exit Codes

- `0` - Success
- `1` - Error occurred

**Examples:**

```bash
# Success
$ tsunami 3000 -f
$ echo $?
0

# Failure (no process)
$ tsunami 3000 -f
Error: no process listening on port 3000
$ echo $?
1

# Failure (invalid port)
$ tsunami 99999
Error: invalid port: 99999 (must be 1-65535)
$ echo $?
1
```

## Environment Variables

Tsunami does not currently use environment variables. All configuration is through command-line flags.

## Interactive TUI Controls

When running in interactive mode (`tsunami` with no arguments):

| Key | Action |
|-----|--------|
| **Type** | Start/continue filtering |
| **↑ / k** | Move selection up |
| **↓ / j** | Move selection down |
| **Enter** | Select process to kill |
| **Backspace** | Delete last filter character |
| **Esc** | Clear filter (or quit if filter is empty) |
| **Ctrl+C** | Quit immediately |

## Examples

### Basic Usage

```bash
# Interactive mode
tsunami

# List all ports
tsunami -l

# Kill single port
tsunami 3000

# Kill without confirmation
tsunami 3000 -f
```

### Multiple Ports

```bash
# Space-separated
tsunami 3000 8080 5432 -f

# Comma-separated
tsunami 3000,8080,5432 -f

# Range
tsunami 3000-3010 -f

# Mixed
tsunami 3000 8080-8090 9000,9001 -f
```

### Signals

```bash
# Default (SIGTERM with escalation)
tsunami 3000

# Immediate SIGKILL
tsunami 3000 -s KILL -f

# SIGINT (Ctrl+C)
tsunami 3000 -s INT

# Custom timeout
tsunami 3000 --timeout 10s
```

### Filtering

```bash
# By process name
tsunami -l --filter node

# By user
tsunami -l --filter user=postgres

# JSON output
tsunami -l --filter node --json
```

### Advanced

```bash
# Dry run
tsunami 3000-4000 --dry-run

# Kill all processes on port
tsunami 8080 --all -f

# Kill by PID
tsunami --pid 12345 --pid 12346 -f

# Quiet script usage
tsunami 3000 8080 -f -q || true
```

## See Also

- [Signal Types](./signals.md) - Detailed signal behavior
- [Configuration Options](./configuration.md) - Configuration details
- [Basic Usage Guide](../guides/basic-usage.md) - Usage patterns
- [Common Use Cases](../guides/common-use-cases.md) - Real-world examples
