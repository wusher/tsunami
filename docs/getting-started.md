# Getting Started

Welcome to Tsunami! This guide will help you install Tsunami and run your first commands.

## Prerequisites

Before installing Tsunami, make sure you have:

- **Go 1.21 or later** (if installing from source or via `go install`)
- **macOS or Linux** operating system
- **Terminal access** with appropriate permissions

:::note Platform Requirements
- **macOS**: Tsunami uses `lsof` (pre-installed)
- **Linux**: Tsunami uses `/proc/net/tcp` (kernel built-in)
:::

## Installation

### Option 1: Install via Go (Recommended)

The easiest way to install Tsunami is using `go install`:

```bash
go install github.com/wusher/tsunami/cmd/tsunami@latest
```

This will install the `tsunami` binary to your `$GOPATH/bin` directory (typically `~/go/bin`).

:::tip
Make sure `$GOPATH/bin` is in your `PATH`:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```
:::

### Option 2: Build from Source

Clone the repository and build manually:

```bash
# Clone the repository
git clone https://github.com/wusher/tsunami.git
cd tsunami

# Build the binary
go build -o tsunami ./cmd/tsunami/

# Optionally, move to your PATH
sudo mv tsunami /usr/local/bin/
```

### Option 3: Download Pre-built Binary

Check the [GitHub Releases](https://github.com/wusher/tsunami/releases) page for pre-built binaries for your platform.

## Verify Installation

After installation, verify that Tsunami is working:

```bash
tsunami --version
```

You should see output showing the version number.

## Your First Command

Let's start with a simple example. First, let's see what ports are currently listening on your system:

```bash
tsunami -l
```

You should see output like this:

```
PORT     PID        PROCESS              USER            PROTO
-----------------------------------------------------------------
3000     12345      node                 youruser        tcp
8080     12346      python3              youruser        tcp
5432     12347      postgres             postgres        tcp
```

Now let's kill a process. If you have a development server running on port 3000:

```bash
# Kill with confirmation prompt
tsunami 3000
```

You'll be prompted:

```
Kill node (PID 12345) on port 3000? [y/N]
```

Type `y` and press Enter. The process will be killed!

:::warning
Be careful when killing processes. Make sure you're targeting the right port!
:::

## Interactive TUI Mode

For a more visual experience, launch Tsunami without any arguments:

```bash
tsunami
```

This opens an interactive terminal UI where you can:

1. **Browse** all listening ports
2. **Filter** by typing (e.g., type "node" to show only Node.js processes)
3. **Navigate** with arrow keys
4. **Select** a process with Enter
5. **Confirm** to kill the process

### TUI Controls

| Key | Action |
|-----|--------|
| **Type** | Filter the list |
| **↑/↓** | Navigate up/down |
| **Enter** | Select process to kill |
| **Backspace** | Delete filter character |
| **Esc** | Clear filter or quit |

## Common First Tasks

### Kill a Development Server

```bash
# Your React/Next.js dev server
tsunami 3000 -f

# Your backend API
tsunami 8080 -f
```

### Free Up Multiple Ports

```bash
# Kill processes on multiple ports
tsunami 3000 8080 9000 -f
```

### Check What's Using a Port

```bash
# List all ports and filter by name
tsunami -l --filter node
```

## Troubleshooting

### "No process listening on port X"

This means no process is currently bound to that port. Use `tsunami -l` to see all listening ports.

### "Permission denied"

You may need elevated permissions to kill processes owned by other users:

```bash
sudo tsunami 3000
```

### "Command not found"

Make sure the Tsunami binary is in your `PATH`:

```bash
# Check if it's installed
which tsunami

# If using go install, add GOPATH/bin to PATH
export PATH=$PATH:$(go env GOPATH)/bin
```

### macOS: "lsof not found"

`lsof` is pre-installed on macOS. If you're getting this error, your system may be misconfigured.

### Linux: Cannot read /proc/net/tcp

Ensure you have read permissions:

```bash
ls -la /proc/net/tcp
```

## Next Steps

Now that you have Tsunami installed and working, explore these guides:

- **[Basic Usage](./guides/basic-usage.md)** - Learn all the command-line options
- **[Common Use Cases](./guides/common-use-cases.md)** - Real-world scenarios
- **[Best Practices](./guides/best-practices.md)** - Tips for effective use

## Getting Help

If you need help, use the built-in help command:

```bash
# General help
tsunami --help

# View all flags
tsunami -h
```

For issues or questions, visit the [GitHub Issues](https://github.com/wusher/tsunami/issues) page.

:::tip Quick Reference
Keep this handy:
```bash
tsunami         # Interactive TUI
tsunami 3000    # Kill port 3000
tsunami -l      # List all ports
tsunami --help  # Show help
```
:::
