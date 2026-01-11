# Configuration Options

Tsunami is designed as a simple CLI tool with minimal configuration. All options are provided via command-line flags.

## Configuration Philosophy

Tsunami follows these principles:

1. **No config files** - All configuration via CLI flags
2. **Sensible defaults** - Works out of the box
3. **Explicit behavior** - No hidden settings or environment variables
4. **Stateless** - Each invocation is independent

## Command-Line Configuration

All configuration is done through command-line flags. See [CLI Commands Reference](./cli-commands.md) for complete details.

### Core Settings

| Setting | Flag | Default | Description |
|---------|------|---------|-------------|
| **Force** | `-f`, `--force` | `false` | Skip confirmation prompts |
| **Quiet** | `-q`, `--quiet` | `false` | Suppress success output |
| **Signal** | `-s`, `--signal` | `TERM` | Signal to send |
| **Timeout** | `-t`, `--timeout` | `2s` | Escalation timeout |
| **Dry Run** | `-n`, `--dry-run` | `false` | Preview without killing |

### Example Configurations

#### Development Mode

```bash
# Interactive with defaults
tsunami
```

#### CI/CD Mode

```bash
# Force, quiet, fast timeout
tsunami 3000 8080 -f -q --timeout 1s
```

#### Production Mode

```bash
# Longer timeout for graceful shutdown
tsunami 8080 --timeout 10s
```

#### Debug Mode

```bash
# Dry run to see what would happen
tsunami 3000-4000 --dry-run
```

## Environment Variables

Tsunami currently **does not** use environment variables. This is intentional to keep behavior explicit and predictable.

If you need environment-based configuration, use shell variables:

```bash
# In your .env or script
export DEV_PORTS="3000 8080 5432"
export KILL_TIMEOUT="5s"

# Use in commands
tsunami $DEV_PORTS -f --timeout $KILL_TIMEOUT
```

## Shell Integration

### Aliases

Create aliases for common configurations in your `.bashrc` or `.zshrc`:

```bash
# Quick dev port cleanup
alias killdev='tsunami 3000 8080 5432 -f -q'

# List with nice formatting
alias ports='tsunami -l'

# Force kill with SIGKILL
alias killforce='tsunami -s KILL -f'

# Graceful with long timeout
alias killgrace='tsunami --timeout 10s'
```

### Functions

More complex logic with shell functions:

```bash
# Kill with confirmation
function killport() {
  local port=$1
  echo "About to kill port $port..."
  tsunami "$port"
}

# Kill and restart command
function killrestart() {
  local port=$1
  shift
  tsunami "$port" -f && "$@"
}

# Usage: killrestart 3000 npm run dev
```

### Completion

Enable shell completion for better UX:

```bash
# Bash
tsunami completion bash > /etc/bash_completion.d/tsunami

# Zsh
tsunami completion zsh > "${fpath[1]}/_tsunami"

# Fish
tsunami completion fish > ~/.config/fish/completions/tsunami.fish
```

## Project-Level Configuration

### Makefile

```makefile
# Makefile

# Configuration
PORTS := 3000 8080 5432
TIMEOUT := 5s

.PHONY: clean-ports
clean-ports:
	@tsunami $(PORTS) -f -q --timeout $(TIMEOUT) || true

.PHONY: dev
dev: clean-ports
	npm run dev
```

### Package.json Scripts

```json
{
  "scripts": {
    "predev": "tsunami 3000 8080 -f -q || true",
    "dev": "next dev",
    "clean": "tsunami 3000 8080 5432 6379 -f",
    "clean:force": "tsunami 3000 8080 5432 6379 -s KILL -f"
  }
}
```

### Shell Scripts

```bash
#!/bin/bash
# cleanup-ports.sh

# Configuration
readonly PORTS="3000 8080 5432"
readonly TIMEOUT="5s"
readonly FORCE=true

# Build command
CMD="tsunami $PORTS"
[ "$FORCE" = true ] && CMD="$CMD -f"
CMD="$CMD --timeout $TIMEOUT"

# Execute
echo "Cleaning ports: $PORTS"
eval "$CMD"
```

## Docker Integration

### docker-compose.yml

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "3000:3000"
    entrypoint:
      - /bin/sh
      - -c
      - |
        # Cleanup before start
        tsunami 3000 -f -q || true
        exec npm start
```

### Dockerfile

```dockerfile
FROM golang:1.21

# Install tsunami
RUN go install github.com/wusher/tsunami/cmd/tsunami@latest

# Use in entrypoint
ENTRYPOINT ["sh", "-c", "tsunami 3000 -f -q || true && exec $0 $@"]
```

## CI/CD Integration

### GitHub Actions

```yaml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Install Tsunami
        run: go install github.com/wusher/tsunami/cmd/tsunami@latest

      - name: Run tests
        run: |
          # Clean ports before tests
          tsunami 3000-3100 -f -q || true

          # Run tests
          npm test

          # Clean ports after tests
          tsunami 3000-3100 -f -q || true
```

### GitLab CI

```yaml
test:
  stage: test
  before_script:
    - go install github.com/wusher/tsunami/cmd/tsunami@latest
    - export PATH=$PATH:$(go env GOPATH)/bin
  script:
    - tsunami 3000 8080 -f -q || true
    - npm test
  after_script:
    - tsunami 3000 8080 -f -q || true
```

## Platform-Specific Behavior

### macOS

Tsunami uses `lsof` on macOS:

```bash
# lsof is pre-installed
tsunami -l
```

**Configuration:** None needed, works out of the box.

### Linux

Tsunami uses `/proc/net/tcp` on Linux:

```bash
# Reads from /proc filesystem
tsunami -l
```

**Configuration:** None needed, kernel built-in.

**Permissions:** Read access to `/proc/net/tcp` required (normally available).

## Logging Configuration

Tsunami doesn't have built-in logging, but you can redirect output:

```bash
# Log to file
tsunami 3000 -f 2>&1 | tee tsunami.log

# Log errors only
tsunami 3000 -f -q 2>> errors.log

# Structured logging with timestamp
tsunami 3000 -f | while read line; do
  echo "$(date -Iseconds) $line"
done >> tsunami.log
```

## Advanced Patterns

### Conditional Configuration

```bash
# Different timeouts based on environment
if [ "$ENVIRONMENT" = "production" ]; then
  TIMEOUT="10s"
else
  TIMEOUT="2s"
fi

tsunami 8080 --timeout $TIMEOUT -f
```

### Dynamic Port Lists

```bash
# Read ports from file
PORTS=$(cat .ports | grep -v "^#" | xargs)
tsunami $PORTS -f -q
```

### Configuration Validation

```bash
#!/bin/bash

validate_port() {
  local port=$1
  if [ "$port" -lt 1 ] || [ "$port" -gt 65535 ]; then
    echo "Invalid port: $port"
    exit 1
  fi
}

# Validate before using
for port in $PORTS; do
  validate_port "$port"
done

tsunami $PORTS -f
```

## No Global Configuration

Tsunami deliberately **does not support**:

- Global config files (`~/.tsunamirc`)
- System-wide defaults (`/etc/tsunami.conf`)
- Environment variable configuration
- Per-user preferences

**Rationale:**
- Explicit is better than implicit
- Behavior is predictable
- No hidden state
- Each command is self-contained

If you need consistent configuration, use:
- Shell aliases
- Shell functions
- Project scripts
- Makefile targets
- Package.json scripts

## FAQ

### Can I set default flags?

Use an alias:

```bash
alias tsunami='command tsunami -f -q'
```

### Can I configure per-project settings?

Yes, via scripts or Makefile:

```bash
# scripts/tsunami.sh
#!/bin/bash
tsunami 3000 8080 -f --timeout 5s "$@"
```

### Can I use config files?

Not directly, but you can source them in scripts:

```bash
# config.sh
PORTS="3000 8080"
TIMEOUT="5s"

# cleanup.sh
source config.sh
tsunami $PORTS --timeout $TIMEOUT -f
```

### How do I share configuration across team?

Commit scripts to version control:

```bash
git add scripts/cleanup-ports.sh
git commit -m "Add port cleanup script"
```

## See Also

- [CLI Commands Reference](./cli-commands.md) - All command-line options
- [Best Practices](../guides/best-practices.md) - Configuration patterns
- [Examples](../examples/development-workflows.md) - Real-world configurations
