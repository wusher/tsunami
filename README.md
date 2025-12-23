# Tsunami

Kill processes listening on ports.

## Install

```bash
go install github.com/wusher/tsunami/cmd/tsunami@latest
```

Or build from source:

```bash
go build -o tsunami ./cmd/tsunami/
```

## Usage

```bash
# Interactive TUI - browse and kill
tsunami

# Kill process on port 3000 (prompts for confirmation)
tsunami 3000

# Kill without confirmation
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
| `--signal` | `-s` | Signal to send (TERM, KILL, INT). Default: TERM |
| `--list` | `-l` | List listening ports and exit |
| `--quiet` | `-q` | Suppress output except errors |

## TUI Controls

| Key | Action |
|-----|--------|
| Type | Filter list |
| Up/Down | Navigate |
| Enter | Select process to kill |
| Backspace | Delete filter character |
| Esc | Clear filter / Quit |

## Platform Support

- macOS (via `lsof`)
- Linux (via `/proc/net/tcp`)

## License

MIT
