# Basic Usage

This guide covers all the basic ways to use Tsunami for killing processes on ports.

## Command Structure

The basic command structure is:

```bash
tsunami [port...] [flags]
```

## Three Modes of Operation

### 1. Interactive TUI Mode

Launch without arguments to browse and kill interactively:

```bash
tsunami
```

**Features:**
- Visual list of all listening ports
- Real-time filtering
- Arrow key navigation
- Safe confirmation before killing

**When to use:** When you're not sure which port or want to browse multiple processes.

### 2. Direct Port Mode

Target specific port(s) directly:

```bash
tsunami 3000
```

**When to use:** When you know exactly which port to kill.

### 3. List Mode

View all listening ports without killing:

```bash
tsunami -l
```

**When to use:** When you want to inspect what's running without taking action.

## Port Specification

### Single Port

Kill a process on a single port:

```bash
tsunami 3000
```

### Multiple Ports

Kill processes on multiple ports:

```bash
# Space-separated
tsunami 3000 8080 5432

# Comma-separated
tsunami 3000,8080,5432
```

### Port Ranges

Kill all processes in a port range:

```bash
# Kill ports 3000 through 3010
tsunami 3000-3010
```

:::warning Range Limit
Port ranges are limited to 1000 ports maximum to prevent accidental mass killing.
:::

### Mixed Syntax

Combine different port specifications:

```bash
tsunami 3000 8080-8090 9000,9001
```

## Common Flags

### Force Mode (`-f`, `--force`)

Skip confirmation prompts:

```bash
tsunami 3000 -f
```

**Use case:** Automation, scripts, when you're certain.

### List Mode (`-l`, `--list`)

List all listening ports:

```bash
tsunami -l
```

**Output:**
```
PORT     PID        PROCESS              USER            PROTO
-----------------------------------------------------------------
3000     12345      node                 youruser        tcp
8080     12346      python3              youruser        tcp
```

### Quiet Mode (`-q`, `--quiet`)

Suppress success output (errors still shown):

```bash
tsunami 3000 -f -q
```

**Use case:** Scripts where you only want to see errors.

### Dry Run (`-n`, `--dry-run`)

Preview what would be killed without actually killing:

```bash
tsunami 3000 --dry-run
```

**Output:**
```
Would kill: node (PID 12345) on port 3000 with signal TERM
```

### Signal Selection (`-s`, `--signal`)

Choose which signal to send:

```bash
# Send SIGKILL immediately (no graceful shutdown)
tsunami 3000 -s KILL

# Send SIGINT (Ctrl+C)
tsunami 3000 -s INT

# Send SIGHUP (reload)
tsunami 3000 -s HUP
```

**Available signals:** `TERM` (default), `KILL`, `INT`, `HUP`

### Filter (`--filter`)

Filter ports by process name or user (list mode only):

```bash
# Show only node processes
tsunami -l --filter node

# Show only processes by specific user
tsunami -l --filter user=postgres
```

### JSON Output (`--json`)

Output in JSON format (list mode only):

```bash
tsunami -l --json
```

**Output:**
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

## Advanced Flags

### Kill All (`-a`, `--all`)

When multiple processes listen on the same port, kill all of them:

```bash
tsunami 8080 --all -f
```

### Timeout (`-t`, `--timeout`)

Customize the escalation timeout (TERM → KILL):

```bash
# Wait 5 seconds before escalating
tsunami 3000 --timeout 5s
```

Default is 2 seconds.

### Kill by PID (`-p`, `--pid`)

Kill processes directly by PID instead of port:

```bash
# Single PID
tsunami --pid 12345

# Multiple PIDs
tsunami --pid 12345 --pid 12346
```

## Combining Flags

Flags can be combined for powerful operations:

```bash
# Kill multiple ports forcefully and quietly
tsunami 3000 8080 9000 -f -q

# List node processes in JSON
tsunami -l --filter node --json

# Dry run with custom signal
tsunami 3000 --dry-run -s KILL

# Kill with custom timeout
tsunami 3000 --timeout 10s -f
```

## Exit Codes

Tsunami uses standard exit codes:

- `0` - Success
- `1` - Error occurred

**Use in scripts:**
```bash
if tsunami 3000 -f; then
  echo "Port 3000 freed successfully"
else
  echo "Failed to free port 3000"
fi
```

## Shell Completion

Generate shell completion for your shell:

```bash
# Bash
tsunami completion bash > /etc/bash_completion.d/tsunami

# Zsh
tsunami completion zsh > "${fpath[1]}/_tsunami"

# Fish
tsunami completion fish > ~/.config/fish/completions/tsunami.fish
```

## Interactive TUI Details

### Filtering

Start typing to filter the list:

```
Type: node
```

This will show only processes matching "node".

### Navigation

- **↑/↓**: Move up/down
- **Enter**: Select process
- **Backspace**: Remove last filter character
- **Esc**: Clear filter or quit (press twice to quit)

### Selection Flow

1. Navigate to process (or filter to find it)
2. Press **Enter**
3. Confirm on the confirmation screen
4. Process is killed

## Tips & Tricks

### Quick Port Check

```bash
# Check if anything is on port 3000
tsunami -l | grep 3000
```

### Kill All Node Processes on Ports

```bash
# List node processes, extract ports, kill them
tsunami -l --filter node --json | jq -r '.[].port' | xargs -I {} tsunami {} -f
```

### Alias for Common Ports

Add to your `.bashrc` or `.zshrc`:

```bash
alias killdev='tsunami 3000 8080 5432 -f'
```

## Next Steps

- **[Common Use Cases](./common-use-cases.md)** - Real-world scenarios
- **[Best Practices](./best-practices.md)** - Tips for effective use
- **[CLI Reference](../reference/cli-commands.md)** - Complete command reference

:::tip
Use `tsunami --help` to see a quick reference of all flags and options.
:::
