# Best Practices

Follow these best practices to use Tsunami effectively and safely in your development workflow.

## Safety First

### Always Preview with Dry Run

Before killing processes, especially with ranges or wildcards, use dry run:

```bash
# Preview what will be killed
tsunami 3000-4000 --dry-run

# Review output, then execute
tsunami 3000-4000 -f
```

### Use List Mode to Verify

Check what's running before killing:

```bash
# See what's on the port
tsunami -l | grep 3000

# Then kill
tsunami 3000
```

### Avoid the Nuclear Option

Don't kill all ports indiscriminately:

```bash
# ❌ BAD - Kills everything including system services
tsunami -l --json | jq -r '.[].port' | xargs -I {} tsunami {} -f

# ✅ GOOD - Target specific ports or ranges
tsunami 3000-9000 -f
```

### Be Careful with sudo

Only use `sudo` when necessary:

```bash
# ❌ BAD - Unnecessary privilege escalation
sudo tsunami 3000

# ✅ GOOD - Only when needed for system ports or other users
sudo tsunami 80 443 -f
```

## Development Workflow

### Integrate with Project Scripts

Add Tsunami to your `package.json`:

```json
{
  "scripts": {
    "predev": "tsunami 3000 8080 -f -q || true",
    "dev": "next dev",
    "clean": "tsunami 3000 8080 5432 6379 -f"
  }
}
```

### Create Project-Specific Cleanup Scripts

```bash
# scripts/cleanup-ports.sh
#!/bin/bash
# Cleanup ports for MyProject

echo "Cleaning up MyProject ports..."
tsunami 3000 8080 5432 -f -q

echo "Ports cleaned. Ready to start services."
```

Make it executable:
```bash
chmod +x scripts/cleanup-ports.sh
```

### Use Environment Variables

Make your scripts portable:

```bash
# .env
DEV_PORT=3000
API_PORT=8080
DB_PORT=5432

# cleanup.sh
source .env
tsunami $DEV_PORT $API_PORT $DB_PORT -f -q
```

## Automation

### Always Use Force Flag in Scripts

Never let scripts hang on confirmation prompts:

```bash
# ❌ BAD - Will hang waiting for input
tsunami 3000

# ✅ GOOD - Executes immediately
tsunami 3000 -f
```

### Handle Errors Gracefully

Don't let port cleanup failures break your scripts:

```bash
# Allow script to continue even if no process is on port
tsunami 3000 -f -q || true

# Or handle the error explicitly
if ! tsunami 3000 -f -q; then
  echo "Warning: Could not kill port 3000 (may not be in use)"
fi
```

### Use Quiet Mode in CI/CD

Reduce noise in build logs:

```bash
# In CI pipeline
tsunami 3000 8080 5432 -f -q
```

### Add Cleanup to Pre-commit Hooks

```bash
# .git/hooks/pre-commit
#!/bin/bash
# Cleanup test ports before commit

tsunami 3000-3100 -f -q || true
```

## Performance

### Use Ranges Instead of Multiple Commands

```bash
# ❌ SLOW - Multiple invocations
tsunami 3000 -f
tsunami 3001 -f
tsunami 3002 -f

# ✅ FAST - Single invocation
tsunami 3000-3002 -f
```

### Filter Early

When searching for specific processes:

```bash
# ❌ INEFFICIENT - Gets all ports then greps
tsunami -l | grep node

# ✅ EFFICIENT - Filters at source
tsunami -l --filter node
```

### Use JSON for Parsing

When processing output programmatically:

```bash
# ❌ FRAGILE - Parsing text table
tsunami -l | awk '{print $1}'

# ✅ ROBUST - Using JSON
tsunami -l --json | jq -r '.[].port'
```

## Monitoring & Logging

### Log Port Operations in Scripts

```bash
#!/bin/bash
LOGFILE="port-cleanup.log"

echo "$(date): Cleaning ports 3000-3010" >> $LOGFILE
tsunami 3000-3010 -f -q >> $LOGFILE 2>&1

if [ $? -eq 0 ]; then
  echo "$(date): ✓ Success" >> $LOGFILE
else
  echo "$(date): ✗ Failed" >> $LOGFILE
fi
```

### Create Audit Trails

```bash
# audit-kill.sh
#!/bin/bash

PORT=$1
USER=$(whoami)
TIMESTAMP=$(date -Iseconds)

echo "$TIMESTAMP,$USER,$PORT,$(tsunami -l --json | jq ".[] | select(.port == $PORT)")" >> port-audit.log

tsunami $PORT -f
```

## Security

### Never Run as Root Unless Necessary

```bash
# ❌ BAD - Unnecessary root
sudo tsunami 3000

# ✅ GOOD - Only your processes
tsunami 3000
```

### Validate Input in Scripts

```bash
#!/bin/bash

PORT=$1

# Validate port number
if ! [[ "$PORT" =~ ^[0-9]+$ ]] || [ "$PORT" -lt 1 ] || [ "$PORT" -gt 65535 ]; then
  echo "Error: Invalid port number"
  exit 1
fi

tsunami $PORT -f
```

### Use User Filters for Shared Systems

```bash
# Only kill your own processes
tsunami -l --filter user=$(whoami) --json | \
  jq -r '.[].port' | \
  xargs -I {} tsunami {} -f
```

## Signal Handling

### Default to SIGTERM

Always prefer graceful shutdown:

```bash
# ✅ GOOD - Allows cleanup
tsunami 3000

# ❌ BAD - Forceful, no cleanup
tsunami 3000 -s KILL -f
```

### Use SIGKILL as Last Resort

Only use SIGKILL when SIGTERM fails:

```bash
# Try graceful first
tsunami 3000 -f

# Wait a moment
sleep 2

# Check if still alive, then force kill
if tsunami -l | grep -q "3000"; then
  echo "Process didn't die, using SIGKILL"
  tsunami 3000 -s KILL -f
fi
```

### Adjust Timeout for Slow Services

Some services need time to shut down gracefully:

```bash
# Database with active connections
tsunami 5432 --timeout 10s

# Quick service
tsunami 3000 --timeout 1s
```

## Documentation

### Comment Your Automation

```bash
# Kill frontend dev server (React on :3000)
tsunami 3000 -f

# Kill backend API (Express on :8080)
tsunami 8080 -f

# Kill database (Postgres on :5432)
tsunami 5432 -f
```

### Create a Project README Section

Document port usage in your README:

```markdown
## Port Management

This project uses the following ports:
- 3000: React frontend
- 8080: Node.js API
- 5432: PostgreSQL database

To clean up all ports:
\`\`\`bash
npm run clean-ports
# or
tsunami 3000 8080 5432 -f
\`\`\`
```

### Maintain a Ports File

Create a `.ports` file in your project:

```
# .ports - Development ports for this project
3000  # Frontend (React)
8080  # Backend API (Express)
5432  # PostgreSQL
6379  # Redis
```

Read it in scripts:

```bash
# cleanup.sh
PORTS=$(cat .ports | grep -v "^#" | awk '{print $1}' | xargs)
tsunami $PORTS -f -q
```

## Testing

### Clean Ports Before and After Tests

```bash
# test-runner.sh
#!/bin/bash

# Clean before
tsunami 3000-3100 -f -q

# Run tests
npm test
EXIT_CODE=$?

# Clean after (even if tests failed)
tsunami 3000-3100 -f -q

exit $EXIT_CODE
```

### Use Different Port Ranges for Different Test Suites

```bash
# Unit tests: 3000-3099
# Integration tests: 3100-3199
# E2E tests: 3200-3299

# Before integration tests
tsunami 3100-3199 -f -q
```

## Aliases & Shortcuts

### Create Useful Aliases

Add to `.bashrc` or `.zshrc`:

```bash
# Common dev ports
alias killdev='tsunami 3000 8080 5432 6379 -f'

# List my ports
alias myports='tsunami -l --filter user=$(whoami)'

# Emergency kill all
alias killall-ports='tsunami 1024-49151 -f -q'

# Project-specific
alias killmyapp='tsunami 3000 8080 -f && echo "MyApp ports cleaned"'
```

### Create Shell Functions

```bash
# Kill port and restart command
killrestart() {
  local port=$1
  shift
  tsunami $port -f && "$@"
}

# Usage:
# killrestart 3000 npm run dev
```

## Error Handling

### Check Exit Codes

```bash
if tsunami 3000 -f; then
  echo "✓ Port 3000 freed"
  npm run dev
else
  echo "✗ Failed to free port 3000"
  exit 1
fi
```

### Provide User Feedback

```bash
#!/bin/bash

echo "Cleaning development ports..."

if tsunami 3000 8080 5432 -f -q; then
  echo "✓ All ports cleaned successfully"
  echo "Ready to start development servers"
else
  echo "⚠️  Warning: Some ports could not be cleaned"
  echo "Run 'tsunami -l' to see what's still running"
  exit 1
fi
```

## Version Control

### Don't Commit Port-Specific Values

Use environment variables or config files:

```bash
# ❌ BAD - Hardcoded in version control
tsunami 3000 -f

# ✅ GOOD - From environment
tsunami $DEV_PORT -f
```

### Add Cleanup Scripts to Git

```bash
git add scripts/cleanup-ports.sh
git commit -m "Add port cleanup script for development"
```

## Summary Checklist

- [ ] Use `--dry-run` for unfamiliar operations
- [ ] Use `-f` in all automation scripts
- [ ] Use `-q` in CI/CD pipelines
- [ ] Prefer SIGTERM over SIGKILL
- [ ] Document ports in project README
- [ ] Add cleanup to project scripts
- [ ] Handle errors gracefully (use `|| true`)
- [ ] Validate inputs in custom scripts
- [ ] Create project-specific aliases
- [ ] Log important operations
- [ ] Test scripts before committing

## Next Steps

- **[CLI Reference](../reference/cli-commands.md)** - Complete command documentation
- **[Advanced Patterns](../examples/advanced-patterns.md)** - Complex automation scenarios
- **[Development Workflows](../examples/development-workflows.md)** - Real-world examples
