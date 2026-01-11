# Tsunami

**Tsunami** is a powerful CLI tool for killing processes listening on network ports. Built for developers who need quick, reliable control over port management during development and debugging.

## Why Tsunami?

- **🎯 Fast & Intuitive**: Kill processes by port number in seconds
- **📊 Interactive TUI**: Browse all listening ports and kill with a keystroke
- **🔍 Smart Discovery**: Automatically finds processes on any port
- **⚡ Batch Operations**: Handle multiple ports at once with ranges and lists
- **🛡️ Safe by Default**: Confirmation prompts and graceful SIGTERM with escalation
- **🎨 Beautiful Output**: Clean tables, JSON support, and filtering options

## Quick Links

### Getting Started
- [Installation & Setup](./getting-started.md)
- [Basic Usage Guide](./guides/basic-usage.md)
- [Common Use Cases](./guides/common-use-cases.md)

### Reference
- [CLI Commands](./reference/cli-commands.md)
- [Configuration Options](./reference/configuration.md)
- [Signal Types](./reference/signals.md)

### Examples
- [Development Workflows](./examples/development-workflows.md)
- [Scripting & Automation](./examples/scripting.md)
- [Advanced Patterns](./examples/advanced-patterns.md)

## Key Features

### Interactive TUI Mode
Launch Tsunami without arguments to browse all listening ports in a beautiful terminal interface. Filter, navigate, and kill processes with ease.

```bash
tsunami
```

### Direct Port Targeting
Kill processes on specific ports instantly:

```bash
# Single port with confirmation
tsunami 3000

# Force kill without confirmation
tsunami 3000 -f

# Multiple ports
tsunami 3000 8080 5432

# Port ranges
tsunami 3000-3010
```

### List & Filter
View all listening ports and filter by process name or user:

```bash
# List all ports
tsunami -l

# Filter by process name
tsunami -l --filter node

# JSON output for scripting
tsunami -l --json
```

### Advanced Control
Fine-tune signal handling and process selection:

```bash
# Send specific signal
tsunami 3000 -s KILL

# Custom escalation timeout
tsunami 3000 --timeout 5s

# Kill by PID
tsunami --pid 1234

# Dry run
tsunami 3000 --dry-run
```

## Platform Support

- **macOS**: Uses `lsof` for port discovery
- **Linux**: Uses `/proc/net/tcp` for port discovery

:::note
Tsunami requires appropriate permissions to kill processes. You may need `sudo` for processes owned by other users.
:::

## Get Started

Ready to take control of your ports? Head over to the [Getting Started guide](./getting-started.md) to install Tsunami and run your first command.

## Community & Support

- **GitHub**: [github.com/wusher/tsunami](https://github.com/wusher/tsunami)
- **Issues**: Report bugs or request features
- **License**: MIT

:::tip
Use `tsunami --help` anytime to see all available commands and options.
:::
