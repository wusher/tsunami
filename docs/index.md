<p align="center">
  <img src="logo.png" alt="Tsunami Logo" width="200">
</p>

# Tsunami

Kill processes listening on ports. Fast, simple, interactive.

## Features

- **Interactive TUI** - Browse and kill processes visually
- **Direct port targeting** - Kill by port number instantly
- **Batch operations** - Multiple ports, ranges, comma-separated
- **Graceful shutdown** - SIGTERM with automatic escalation to SIGKILL
- **Cross-platform** - macOS and Linux support

## Quick Start

```bash
# Install
go install github.com/wusher/tsunami/cmd/tsunami@latest

# Interactive mode
tsunami

# Kill port 3000
tsunami 3000 -f

# List all listening ports
tsunami -l
```

## Navigation

- [Getting Started](./getting-started.md) - Installation and setup
- [Usage](./usage.md) - Complete CLI reference

## Links

- [GitHub](https://github.com/wusher/tsunami)
- [Releases](https://github.com/wusher/tsunami/releases)
