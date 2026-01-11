# Development Workflows

Real-world examples of integrating Tsunami into your development workflow.

## Web Development

### React/Next.js Development

```bash
#!/bin/bash
# start-dev.sh - Start React development server

# Kill any existing dev server
tsunami 3000 -f -q || true

# Start fresh
echo "Starting React dev server on port 3000..."
npm run dev
```

**With hot reload protection:**

```bash
#!/bin/bash
# safe-restart-dev.sh

PORT=3000

# Check if port is in use
if tsunami -l --json | jq -e ".[] | select(.port == $PORT)" > /dev/null; then
  echo "⚠️  Port $PORT is in use"

  # Show what's using it
  tsunami -l | grep $PORT

  # Ask to kill
  read -p "Kill and restart? [y/N] " -n 1 -r
  echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    tsunami $PORT -f
  else
    exit 1
  fi
fi

echo "Starting development server..."
npm run dev
```

### Full-Stack Development

```bash
#!/bin/bash
# start-fullstack.sh - Start frontend + backend + database

# Configuration
FRONTEND_PORT=3000
BACKEND_PORT=8080
DB_PORT=5432

# Clean all ports
echo "Cleaning ports..."
tsunami $FRONTEND_PORT $BACKEND_PORT $DB_PORT -f -q || true

# Start database
echo "Starting PostgreSQL..."
docker-compose up -d postgres

# Wait for DB to be ready
sleep 2

# Start backend
echo "Starting API server..."
cd backend && npm run dev &

# Start frontend
echo "Starting frontend..."
cd frontend && npm run dev &

# Show status
sleep 2
echo ""
echo "Services started:"
tsunami -l | grep -E "$FRONTEND_PORT|$BACKEND_PORT|$DB_PORT"
```

### Multiple Projects Switcher

```bash
#!/bin/bash
# switch-project.sh - Switch between development projects

# Define projects and their ports
declare -A PROJECTS=(
  ["project-a"]="3000 8080"
  ["project-b"]="3001 8081"
  ["project-c"]="3002 8082"
)

PROJECT=$1

if [ -z "$PROJECT" ] || [ -z "${PROJECTS[$PROJECT]}" ]; then
  echo "Usage: $0 <project>"
  echo "Available projects:"
  for name in "${!PROJECTS[@]}"; do
    echo "  - $name (ports: ${PROJECTS[$name]})"
  done
  exit 1
fi

PORTS=${PROJECTS[$PROJECT]}

# Kill all ports from all projects
echo "Stopping all projects..."
for project_ports in "${PROJECTS[@]}"; do
  tsunami $project_ports -f -q || true
done

# Switch to new project
echo "Switching to $PROJECT..."
cd "$PROJECT" || exit 1

# Start the project
echo "Starting $PROJECT on ports $PORTS..."
npm run dev
```

## Testing Workflows

### Test Runner with Port Cleanup

```bash
#!/bin/bash
# run-tests.sh - Run tests with port cleanup

# Test port range
TEST_PORTS="3000-3100"

# Cleanup before tests
echo "Cleaning test ports..."
tsunami $TEST_PORTS -f -q || true

# Run tests
echo "Running tests..."
npm test
TEST_EXIT_CODE=$?

# Cleanup after tests (even if they failed)
echo "Cleaning up..."
tsunami $TEST_PORTS -f -q || true

# Report results
if [ $TEST_EXIT_CODE -eq 0 ]; then
  echo "✓ Tests passed"
else
  echo "✗ Tests failed"
fi

exit $TEST_EXIT_CODE
```

### Parallel Test Runner

```bash
#!/bin/bash
# parallel-tests.sh - Run test suites in parallel

# Port ranges for each test suite
UNIT_TESTS_PORTS="3000-3019"
INTEGRATION_TESTS_PORTS="3020-3039"
E2E_TESTS_PORTS="3040-3059"

# Cleanup all test ports
echo "Cleaning all test ports..."
tsunami 3000-3059 -f -q || true

# Run in parallel
echo "Running tests in parallel..."

npm run test:unit & UNIT_PID=$!
npm run test:integration & INTEGRATION_PID=$!
npm run test:e2e & E2E_PID=$!

# Wait for all to complete
wait $UNIT_PID
UNIT_EXIT=$?

wait $INTEGRATION_PID
INTEGRATION_EXIT=$?

wait $E2E_PID
E2E_EXIT=$?

# Cleanup
tsunami 3000-3059 -f -q || true

# Report
echo ""
echo "Test Results:"
echo "  Unit: $([ $UNIT_EXIT -eq 0 ] && echo '✓' || echo '✗')"
echo "  Integration: $([ $INTEGRATION_EXIT -eq 0 ] && echo '✓' || echo '✗')"
echo "  E2E: $([ $E2E_EXIT -eq 0 ] && echo '✓' || echo '✗')"

# Exit with failure if any suite failed
[ $UNIT_EXIT -eq 0 ] && [ $INTEGRATION_EXIT -eq 0 ] && [ $E2E_EXIT -eq 0 ]
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit - Clean test ports before commit

# Clean test ports
tsunami 3000-3100 -f -q || true

# Run quick tests
npm run test:quick

exit $?
```

## Docker Workflows

### Docker Compose with Port Cleanup

```bash
#!/bin/bash
# docker-start.sh - Start Docker services with port cleanup

# Ports used by docker-compose
PORTS="3000 5432 6379 8080"

# Stop any existing services
echo "Stopping existing services..."
docker-compose down 2>/dev/null || true

# Clean ports
echo "Cleaning ports..."
tsunami $PORTS -f -q || true

# Start services
echo "Starting Docker services..."
docker-compose up -d

# Wait for services to be ready
echo "Waiting for services to start..."
sleep 3

# Show status
echo ""
echo "Services status:"
docker-compose ps
echo ""
echo "Port usage:"
tsunami -l | grep -E "$(echo $PORTS | tr ' ' '|')"
```

### Cleanup Orphaned Containers

```bash
#!/bin/bash
# cleanup-docker-ports.sh - Clean ports from stopped containers

# Get ports from docker-compose
PORTS=$(docker-compose ps --all --format json 2>/dev/null | \
        jq -r '.[].Publishers[]?.PublishedPort' 2>/dev/null | \
        sort -u | tr '\n' ' ')

if [ -z "$PORTS" ]; then
  echo "No Docker ports found"
  exit 0
fi

echo "Cleaning Docker ports: $PORTS"
tsunami $PORTS -f -q || true

echo "Removing stopped containers..."
docker-compose down
```

## CI/CD Workflows

### GitHub Actions Integration

```bash
#!/bin/bash
# ci-test.sh - CI test script with port cleanup

set -e  # Exit on error

# CI environment detection
if [ -n "$CI" ]; then
  echo "Running in CI environment"
  FORCE="-f -q"
else
  echo "Running locally"
  FORCE="-f"
fi

# Cleanup
tsunami 3000-4000 $FORCE || true

# Run linter
echo "Running linter..."
npm run lint

# Run tests
echo "Running tests..."
npm test

# Cleanup again
tsunami 3000-4000 $FORCE || true

echo "✓ All checks passed"
```

### Deploy Script

```bash
#!/bin/bash
# deploy.sh - Deploy with port cleanup

set -e

ENV=${1:-staging}

echo "Deploying to $ENV..."

# Cleanup deployment ports
case $ENV in
  staging)
    tsunami 8080 -f -q || true
    ;;
  production)
    tsunami 80 443 -f --timeout 10s || true
    ;;
esac

# Build
echo "Building..."
npm run build

# Deploy
echo "Deploying..."
./deploy-$ENV.sh

echo "✓ Deployed to $ENV"
```

## Makefile Integration

```makefile
# Makefile for development workflow

# Configuration
PORTS := 3000 8080 5432 6379
TIMEOUT := 5s

.PHONY: clean-ports
clean-ports:
	@echo "Cleaning ports: $(PORTS)"
	@tsunami $(PORTS) -f -q --timeout $(TIMEOUT) || true

.PHONY: dev
dev: clean-ports
	@echo "Starting development servers..."
	@npm run dev

.PHONY: test
test: clean-ports
	@echo "Running tests..."
	@npm test
	@$(MAKE) clean-ports

.PHONY: docker-up
docker-up: clean-ports
	@echo "Starting Docker services..."
	@docker-compose up -d

.PHONY: docker-down
docker-down:
	@docker-compose down
	@$(MAKE) clean-ports

.PHONY: restart
restart: docker-down docker-up

.PHONY: status
status:
	@echo "Port status:"
	@tsunami -l | grep -E "$(shell echo $(PORTS) | tr ' ' '|')" || echo "No services running"
```

Usage:

```bash
make dev          # Clean ports and start dev
make test         # Run tests with cleanup
make docker-up    # Start Docker services
make restart      # Restart everything
make status       # Show port status
```

## Package.json Integration

```json
{
  "name": "my-app",
  "scripts": {
    "predev": "tsunami 3000 8080 -f -q || true",
    "dev": "next dev",
    "postdev": "tsunami 3000 8080 -f -q || true",

    "pretest": "tsunami 3000-3100 -f -q || true",
    "test": "jest",
    "posttest": "tsunami 3000-3100 -f -q || true",

    "clean": "tsunami 3000 8080 5432 6379 -f -q",
    "clean:force": "tsunami 3000 8080 5432 6379 -s KILL -f -q",

    "ports": "tsunami -l",
    "ports:mine": "tsunami -l --filter user=$USER",

    "start:clean": "npm run clean && npm start",
    "dev:clean": "npm run clean && npm run dev",

    "docker:up": "npm run clean && docker-compose up -d",
    "docker:down": "docker-compose down && npm run clean"
  }
}
```

## Shell Aliases

Add to `.bashrc` or `.zshrc`:

```bash
# Tsunami aliases for development
alias killdev='tsunami 3000 8080 5432 6379 -f -q'
alias ports='tsunami -l'
alias myports='tsunami -l --filter user=$USER'
alias killmine='tsunami -l --filter user=$USER --json | jq -r ".[].port" | xargs -I {} tsunami {} -f'

# Project-specific
alias kill-myapp='tsunami 3000 8080 -f && echo "MyApp ports cleaned"'
alias start-myapp='kill-myapp && cd ~/projects/myapp && npm run dev'

# Docker helpers
alias kill-docker='tsunami 3000 5432 6379 8080 -f -q'
alias restart-docker='docker-compose down && kill-docker && docker-compose up -d'
```

## Advanced Patterns

### Port Monitor

```bash
#!/bin/bash
# monitor-ports.sh - Monitor and auto-restart services

PORTS="3000 8080"
CHECK_INTERVAL=5

while true; do
  for PORT in $PORTS; do
    if ! tsunami -l --json | jq -e ".[] | select(.port == $PORT)" > /dev/null; then
      echo "⚠️  Port $PORT is down! Restarting..."

      case $PORT in
        3000)
          npm run dev &
          ;;
        8080)
          npm run api &
          ;;
      esac
    fi
  done

  sleep $CHECK_INTERVAL
done
```

### Smart Restart

```bash
#!/bin/bash
# smart-restart.sh - Restart only what changed

PORT=$1
SERVICE=$2

# Get current PID and command
CURRENT=$(tsunami -l --json | jq -r ".[] | select(.port == $PORT) | .process")

if [ -z "$CURRENT" ]; then
  echo "No service on port $PORT, starting..."
  eval "$SERVICE" &
else
  echo "Service already running: $CURRENT"
  read -p "Restart? [y/N] " -n 1 -r
  echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    tsunami $PORT -f
    eval "$SERVICE" &
  fi
fi
```

## Next Steps

- **[Scripting & Automation](./scripting.md)** - More automation patterns
- **[Advanced Patterns](./advanced-patterns.md)** - Complex scenarios
- **[Best Practices](../guides/best-practices.md)** - Tips for effective workflows
