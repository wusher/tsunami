# Common Use Cases

Learn how to use Tsunami for real-world development scenarios.

## Development Workflows

### Restart Development Server

When your dev server is stuck or needs a fresh start:

```bash
# Kill and restart
tsunami 3000 -f && npm run dev
```

For multiple services:

```bash
# Kill all dev services and restart
tsunami 3000 8080 5432 -f && docker-compose up
```

### Free Ports Before Starting Services

Ensure ports are free before starting your application:

```bash
#!/bin/bash
# startup.sh

# Free common dev ports
tsunami 3000 8080 5432 -f -q

# Start services
npm run dev &
npm run api &
docker-compose up -d postgres
```

### Clean Slate Development

Reset all development services:

```bash
# Kill common dev ports
tsunami 3000-3010 8000-8090 5432 6379 -f -q

# Restart everything fresh
docker-compose down && docker-compose up -d
npm run dev
```

## CI/CD Pipelines

### Cleanup After Tests

Ensure test servers don't leak:

```bash
# In your CI pipeline
npm test

# Cleanup regardless of test result
tsunami 3000 8080 -f -q || true
```

### Pre-deployment Port Check

```yaml
# .github/workflows/deploy.yml
- name: Clean up ports
  run: |
    tsunami 8080 -f -q || true

- name: Start application
  run: ./start-server.sh
```

### Docker Container Port Conflicts

```bash
# Before starting containers
tsunami 3000 5432 6379 -f -q

# Start docker services
docker-compose up -d
```

## Multiple Environments

### Switch Between Projects

When switching between projects that use the same ports:

```bash
# Stop current project ports
tsunami 3000 8080 -f -q

# Switch to new project
cd ../other-project
npm run dev
```

### Shared Development Machine

Multiple developers or services on one machine:

```bash
# List ports by user
tsunami -l --filter user=$(whoami)

# Kill only your processes
tsunami -l --filter user=$(whoami) --json | \
  jq -r '.[].port' | \
  xargs -I {} tsunami {} -f
```

## Debugging

### Find What's Using a Port

```bash
# Quick check
tsunami -l | grep 3000

# Detailed information
tsunami -l --filter node

# JSON for parsing
tsunami -l --json | jq '.[] | select(.port == 3000)'
```

### Stuck Process Investigation

When a process won't die:

```bash
# Try graceful termination (default SIGTERM)
tsunami 3000

# If that fails, force kill
tsunami 3000 -s KILL -f
```

### Multiple Processes on Same Port

Sometimes multiple processes bind to the same port:

```bash
# See the error
tsunami 8080
# Error: multiple processes on port 8080: 12345, 12346. Use --all to kill all

# Kill all of them
tsunami 8080 --all -f
```

## Scripting & Automation

### Automated Cleanup Script

Create a cleanup script for your project:

```bash
#!/bin/bash
# cleanup-ports.sh

PORTS="3000 8080 8443 5432 6379"

echo "Cleaning up development ports..."
tsunami $PORTS -f -q

if [ $? -eq 0 ]; then
  echo "✓ All ports cleaned"
else
  echo "⚠ Some ports could not be cleaned"
fi
```

### Conditional Port Killing

Only kill if port is in use:

```bash
#!/bin/bash

PORT=3000

if tsunami -l --json | jq -e ".[] | select(.port == $PORT)" > /dev/null; then
  echo "Port $PORT is in use, killing..."
  tsunami $PORT -f
else
  echo "Port $PORT is free"
fi
```

### Integration with Make

```makefile
# Makefile

.PHONY: clean-ports
clean-ports:
	@tsunami 3000 8080 5432 -f -q || true

.PHONY: dev
dev: clean-ports
	npm run dev

.PHONY: test
test: clean-ports
	npm test
	@tsunami 3000 8080 -f -q || true
```

### Integration with Package Scripts

```json
{
  "scripts": {
    "predev": "tsunami 3000 8080 -f -q || true",
    "dev": "next dev",
    "pretest": "tsunami 3000 8080 -f -q || true",
    "test": "jest"
  }
}
```

## Monitoring

### Log Port Usage

Track what's using ports over time:

```bash
#!/bin/bash
# monitor-ports.sh

while true; do
  echo "=== $(date) ==="
  tsunami -l
  echo ""
  sleep 60
done >> port-monitor.log
```

### Alert on Port Usage

```bash
#!/bin/bash
# alert-port.sh

WATCH_PORT=3000

if tsunami -l --json | jq -e ".[] | select(.port == $WATCH_PORT)" > /dev/null; then
  echo "⚠️  WARNING: Port $WATCH_PORT is in use!"
  tsunami -l | grep $WATCH_PORT
fi
```

## Docker Development

### Clean Ports Before Docker Compose

```bash
# docker-start.sh
#!/bin/bash

# Kill ports that docker-compose will use
tsunami 3000 5432 6379 8080 -f -q

# Start services
docker-compose up -d

# Show what's running
docker-compose ps
```

### Port Conflict Resolution

```bash
# When docker fails due to port conflict
docker-compose up -d || {
  echo "Port conflict detected, cleaning up..."
  tsunami 3000 5432 6379 -f
  docker-compose up -d
}
```

## Testing

### Test Isolation

Ensure each test has clean ports:

```bash
# test.sh
#!/bin/bash

# Before test suite
tsunami 3000-3100 -f -q

# Run tests
npm test

# After test suite
tsunami 3000-3100 -f -q
```

### Parallel Test Runners

When running tests in parallel:

```bash
# Each test gets its own port range
# Test runner 1: ports 3000-3010
# Test runner 2: ports 3010-3020

# Cleanup before
tsunami 3000-3020 -f -q

# Run parallel tests
npm run test:parallel

# Cleanup after
tsunami 3000-3020 -f -q
```

## Emergency Scenarios

### Nuclear Option - Kill Everything

When nothing else works:

```bash
# List everything first
tsunami -l

# Kill all listening ports (use with caution!)
tsunami -l --json | jq -r '.[].port' | sort -u | xargs -I {} tsunami {} -f
```

:::warning
This kills ALL listening ports. System services may be affected!
:::

### Hung Process

When a process won't respond to SIGTERM:

```bash
# First attempt - graceful
tsunami 3000

# If it's still there after 5 seconds
tsunami 3000 -s KILL -f
```

### Permission Issues

When you need sudo:

```bash
# Kill process owned by another user
sudo tsunami 80 443 -f
```

## Best Practices

1. **Always use `-f` in scripts** to avoid hanging on prompts
2. **Use `--dry-run` first** when working with ranges or multiple ports
3. **Add cleanup to CI/CD** to prevent port leaks
4. **Use aliases** for frequently used port combinations
5. **Check before killing** in production environments

## Next Steps

- **[Best Practices](./best-practices.md)** - Tips for effective and safe usage
- **[Advanced Patterns](../examples/advanced-patterns.md)** - Complex automation scenarios
- **[CLI Reference](../reference/cli-commands.md)** - Complete command documentation
