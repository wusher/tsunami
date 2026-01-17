# Tsunami

[![CI](https://github.com/wusher/tsunami/actions/workflows/ci.yml/badge.svg)](https://github.com/wusher/tsunami/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Kill processes listening on ports.

## Install

### Pre-built Binaries

Download from [GitHub Releases](https://github.com/wusher/tsunami/releases) for your platform:
- Linux (amd64, arm64)
- macOS (amd64, arm64)

### Go Install

```bash
go install github.com/wusher/tsunami/cmd/tsunami@latest
```

### Build from Source

```bash
go build -o tsunami ./cmd/tsunami/
```

## Usage

```bash
# Interactive TUI - browse and kill
tsunami

# Kill process on port 3000
tsunami 3000 -f

# Kill multiple ports
tsunami 3000 8080 5432 -f

# List listening ports
tsunami -l

# Send specific signal
tsunami 3000 -s KILL
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip confirmation prompt |
| `--signal` | `-s` | Signal to send (TERM, KILL, INT) |
| `--list` | `-l` | List listening ports |
| `--quiet` | `-q` | Suppress output except errors |
| `--dry-run` | `-n` | Show what would be killed |
| `--timeout` | `-t` | Escalation timeout (default: 2s) |

## TUI Controls

| Key | Action |
|-----|--------|
| Type | Filter list |
| Up/Down | Navigate |
| Enter | Select to kill |
| Esc | Quit |

## Platform Support

- macOS (via `lsof`)
- Linux (via `/proc/net/tcp`)

## Documentation

Full documentation: [wusher.github.io/tsunami](https://wusher.github.io/tsunami)

## License

MIT
