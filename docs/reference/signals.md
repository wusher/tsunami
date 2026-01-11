# Signal Types

Understanding Unix signals and how Tsunami uses them to terminate processes.

## Overview

Tsunami uses Unix signals to communicate with processes. Different signals have different behaviors and use cases.

## Supported Signals

### SIGTERM (Default)

**Usage:**
```bash
tsunami 3000
# or explicitly:
tsunami 3000 -s TERM
```

**Behavior:**
1. Sends SIGTERM to the process
2. Waits for timeout period (default: 2s)
3. If process still exists, sends SIGKILL

**Characteristics:**
- **Graceful**: Allows process cleanup
- **Catchable**: Process can handle the signal
- **Escalates**: Automatically upgrades to SIGKILL if needed

**What happens in the process:**
- Cleanup handlers run
- Connections close gracefully
- Files are flushed and closed
- State is saved

**Best for:**
- Normal shutdown
- Development servers
- Processes with cleanup needs
- When you want graceful termination

**Example:**
```bash
# Default graceful kill with 2s timeout
tsunami 3000

# Custom timeout for slow shutdown
tsunami 5432 --timeout 10s
```

### SIGKILL

**Usage:**
```bash
tsunami 3000 -s KILL
```

**Behavior:**
- Immediately terminates the process
- Cannot be caught or ignored
- No cleanup handlers run

**Characteristics:**
- **Immediate**: Process dies instantly
- **Uncatchable**: Cannot be handled or ignored
- **Forceful**: No cleanup possible

**What happens in the process:**
- Nothing - process is killed immediately
- No cleanup code runs
- Connections abruptly close
- May leave orphaned resources

**Best for:**
- Stuck or frozen processes
- When SIGTERM fails
- Emergency situations
- Processes that won't respond

**Example:**
```bash
# Immediate forceful kill
tsunami 3000 -s KILL -f

# After SIGTERM fails
tsunami 3000 -f  # Try graceful first
sleep 2
tsunami 3000 -s KILL -f  # Force if still alive
```

:::warning
SIGKILL should be a last resort. Always try SIGTERM first.
:::

### SIGINT

**Usage:**
```bash
tsunami 3000 -s INT
```

**Behavior:**
- Sends interrupt signal (like pressing Ctrl+C)
- Process can catch and handle it
- No automatic escalation

**Characteristics:**
- **Interactive**: Mimics Ctrl+C
- **Catchable**: Process can handle it
- **Graceful**: Allows cleanup

**What happens in the process:**
- Same as pressing Ctrl+C
- Interrupt handlers run
- Typically results in graceful shutdown

**Best for:**
- Interactive applications
- Processes designed to handle Ctrl+C
- Testing interrupt behavior
- Mimicking manual interruption

**Example:**
```bash
# Send interrupt signal
tsunami 3000 -s INT

# Common for dev servers that handle Ctrl+C well
tsunami 3000-3010 -s INT -f
```

### SIGHUP

**Usage:**
```bash
tsunami 3000 -s HUP
```

**Behavior:**
- Sends "hangup" signal
- Originally meant "terminal disconnected"
- Often used for configuration reload

**Characteristics:**
- **Historical**: Originally for terminal hangup
- **Catchable**: Process can handle it
- **Reload**: Many daemons reload config on SIGHUP

**What happens in the process:**
- Depends on application
- Daemons often reload configuration
- Some processes treat it like SIGTERM

**Best for:**
- Reloading daemon configuration
- Processes that support SIGHUP reload
- Legacy applications

**Example:**
```bash
# Reload nginx config
tsunami 80 -s HUP

# Not typically used for killing
```

:::note
SIGHUP behavior varies by application. Check your process documentation.
:::

## Signal Comparison

| Signal | Catchable | Graceful | Escalates | Speed | Use Case |
|--------|-----------|----------|-----------|-------|----------|
| **TERM** | ✅ | ✅ | ✅ | Medium | Default, graceful shutdown |
| **KILL** | ❌ | ❌ | ❌ | Instant | Stuck processes, emergency |
| **INT** | ✅ | ✅ | ❌ | Fast | Interactive apps, Ctrl+C |
| **HUP** | ✅ | ❓ | ❌ | Fast | Reload config, legacy |

## Escalation Behavior

### SIGTERM Escalation

When using SIGTERM (default), Tsunami automatically escalates:

```bash
tsunami 3000
```

**Timeline:**
1. `T+0s`: Send SIGTERM
2. `T+0s - T+2s`: Wait for process to exit
3. `T+2s`: If still alive, send SIGKILL

**Custom timeout:**
```bash
# Wait 5 seconds before escalating
tsunami 3000 --timeout 5s
```

**Behavior:**
- Checks process every 100ms
- Escalates only if process still exists after timeout
- Returns success if process exits before escalation

### Other Signals

Other signals do NOT escalate:

```bash
# Sends SIGINT once, then returns
tsunami 3000 -s INT

# Sends SIGKILL once (no escalation needed)
tsunami 3000 -s KILL
```

## Process Response

### Well-Behaved Process

```
SIGTERM received → Run cleanup → Exit gracefully → DONE
```

**Example:** Most web servers, databases

```bash
# Process exits before timeout
$ tsunami 3000
Killed node (PID 12345) on port 3000
# Process exited in ~100ms
```

### Slow Shutdown

```
SIGTERM received → Long cleanup → Exit after 5s → DONE
```

**Example:** Database with active connections

```bash
# Need longer timeout
$ tsunami 5432 --timeout 10s
Killed postgres (PID 12347) on port 5432
# Process exited in ~8s
```

### Stuck Process

```
SIGTERM received → Hangs → Timeout → SIGKILL → DEAD
```

**Example:** Deadlocked application

```bash
# Escalates to SIGKILL after timeout
$ tsunami 3000
Killed node (PID 12345) on port 3000
# Process killed via SIGKILL after 2s
```

### Unresponsive Process

```
SIGTERM received → Ignores → Timeout → SIGKILL → DEAD
```

**Example:** Buggy software ignoring signals

```bash
# Direct SIGKILL better here
$ tsunami 3000 -s KILL -f
Killed buggyapp (PID 12348) on port 3000
# Instant kill
```

## Best Practices

### 1. Start with SIGTERM

Always try graceful termination first:

```bash
# ✅ GOOD - Graceful by default
tsunami 3000

# ❌ BAD - Unnecessarily forceful
tsunami 3000 -s KILL
```

### 2. Adjust Timeout for Application Needs

```bash
# Fast services (web servers)
tsunami 3000 --timeout 1s

# Slow services (databases)
tsunami 5432 --timeout 10s

# Very slow (batch processors)
tsunami 8080 --timeout 30s
```

### 3. Use SIGKILL Only When Necessary

```bash
# Try graceful first
if ! tsunami 3000 -f; then
  echo "SIGTERM failed, trying SIGKILL..."
  tsunami 3000 -s KILL -f
fi
```

### 4. Match Signal to Application

```bash
# Web dev server - SIGTERM (default)
tsunami 3000

# Interactive CLI app - SIGINT
tsunami 3000 -s INT

# Daemon config reload - SIGHUP
tsunami 80 -s HUP

# Frozen process - SIGKILL
tsunami 3000 -s KILL
```

### 5. Don't Rely on SIGHUP for Killing

```bash
# ❌ BAD - SIGHUP may not kill
tsunami 3000 -s HUP

# ✅ GOOD - Use SIGTERM for killing
tsunami 3000
```

## Common Scenarios

### Development Server Won't Restart

```bash
# Graceful kill
tsunami 3000 -f && npm run dev
```

### Process Not Responding to SIGTERM

```bash
# Try with longer timeout first
tsunami 3000 --timeout 10s -f

# If still stuck, force kill
tsunami 3000 -s KILL -f
```

### Database Shutdown

```bash
# Long timeout for connection cleanup
tsunami 5432 --timeout 15s
```

### Batch of Processes

```bash
# Range with custom timeout
tsunami 3000-3010 --timeout 5s -f
```

### Emergency Kill

```bash
# Immediate forceful kill
tsunami 3000 -s KILL -f
```

## Signal Numbers

For reference, the signal numbers (may vary by platform):

| Signal | Number | Catchable |
|--------|--------|-----------|
| SIGHUP | 1 | Yes |
| SIGINT | 2 | Yes |
| SIGKILL | 9 | No |
| SIGTERM | 15 | Yes |

:::note
Tsunami uses signal names, not numbers. Use `-s TERM`, not `-s 15`.
:::

## Process States

After signal delivery, processes can be in different states:

### Exiting Gracefully

```bash
$ ps aux | grep myapp
user  12345  0.0  0.0  ... myapp
# Signal sent
# Process running cleanup
$ ps aux | grep myapp
# Process gone
```

### Zombie State

Sometimes processes become zombies (waiting for parent to reap):

```bash
$ ps aux | grep myapp
user  12345  0.0  0.0  ... [myapp] <defunct>
```

Tsunami detects this and considers the process killed.

### Stuck in Uninterruptible Sleep

Rare, but processes can be stuck in kernel operations:

```bash
$ ps aux | grep myapp
user  12345  0.0  0.0  D ... myapp
#                       ^- 'D' state (uninterruptible)
```

Even SIGKILL won't work immediately. Process will exit when kernel operation completes.

## Troubleshooting

### "Process won't die with SIGTERM"

```bash
# Increase timeout
tsunami 3000 --timeout 10s -f

# Still stuck? Use SIGKILL
tsunami 3000 -s KILL -f
```

### "SIGKILL didn't work immediately"

Process may be in uninterruptible sleep. Wait a moment and check:

```bash
tsunami 3000 -s KILL -f
sleep 1
tsunami -l | grep 3000  # Check if still alive
```

### "Process dies but port still shows in use"

The OS may need a moment to release the port:

```bash
tsunami 3000 -f
sleep 1  # Wait for OS to release port
./start-server.sh
```

## See Also

- [CLI Commands Reference](./cli-commands.md) - Full command documentation
- [Best Practices](../guides/best-practices.md) - Safe usage patterns
- [Common Use Cases](../guides/common-use-cases.md) - Real-world examples

## External Resources

- [Linux Signal Man Page](https://man7.org/linux/man-pages/man7/signal.7.html)
- [POSIX Signals](https://pubs.opengroup.org/onlinepubs/9699919799/basedefs/signal.h.html)
