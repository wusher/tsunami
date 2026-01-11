# Tsunami

[![Documentation](https://img.shields.io/badge/docs-latest-blue.svg)](https://wusher.github.io/tsunami)
[![Docs Build](https://github.com/wusher/tsunami/actions/workflows/deploy-docs.yml/badge.svg)](https://github.com/wusher/tsunami/actions/workflows/deploy-docs.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

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

## Documentation

Comprehensive documentation is available at **[wusher.github.io/tsunami](https://wusher.github.io/tsunami)**

### Quick Links

- [Getting Started](https://wusher.github.io/tsunami/getting-started) - Installation and setup
- [Basic Usage](https://wusher.github.io/tsunami/guides/basic-usage) - Command-line options
- [Common Use Cases](https://wusher.github.io/tsunami/guides/common-use-cases) - Real-world scenarios
- [CLI Reference](https://wusher.github.io/tsunami/reference/cli-commands) - Complete command reference

### Building Docs Locally

```bash
# Install Hugo static site generator
# macOS:
brew install hugo

# Linux:
sudo apt-get install hugo
# or: snap install hugo

# Windows:
# Download from https://github.com/gohugoio/hugo/releases

# Build documentation
./scripts/build-docs.sh

# Serve locally at http://localhost:1313
./scripts/serve-docs.sh
```

### Contributing to Documentation

Documentation source files are in the `docs/` folder. To contribute:

1. Edit markdown files in `docs/`
2. Test locally: `./scripts/serve-docs.sh`
3. Submit a pull request

See [SETUP_GITHUB_PAGES.md](docs/SETUP_GITHUB_PAGES.md) for detailed setup information.

## License

MIT
