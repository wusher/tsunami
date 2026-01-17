# Getting Started

## Installation

### Pre-built Binaries (Recommended)

Download from [GitHub Releases](https://github.com/wusher/tsunami/releases) for your platform.

### Go Install

```bash
go install github.com/wusher/tsunami/cmd/tsunami@latest
```

Ensure `$GOPATH/bin` is in your PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Build from Source

```bash
git clone https://github.com/wusher/tsunami.git
cd tsunami
go build -o tsunami ./cmd/tsunami/
```

## Verify Installation

```bash
tsunami --version
```

## First Commands

### List listening ports

```bash
tsunami -l
```

Output:
```
PORT     PID        PROCESS              USER            PROTO
-----------------------------------------------------------------
3000     12345      node                 youruser        tcp
8080     12346      python3              youruser        tcp
```

### Kill a process on a port

```bash
tsunami 3000
```

You'll be prompted for confirmation. Use `-f` to skip:

```bash
tsunami 3000 -f
```

### Interactive mode

```bash
tsunami
```

Use arrow keys to navigate, type to filter, Enter to select.

## Platform Requirements

- **macOS**: Uses `lsof` (pre-installed)
- **Linux**: Uses `/proc/net/tcp` (kernel built-in)

:::note
You may need `sudo` to kill processes owned by other users.
:::

## Next

See [Usage](./usage.md) for the complete CLI reference.
