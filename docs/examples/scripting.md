# Scripting & Automation

Learn how to integrate Tsunami into scripts and automated workflows.

## Basic Script Patterns

### Simple Port Killer Script

```bash
#!/bin/bash
# kill-port.sh - Kill a port with error handling

PORT=$1

if [ -z "$PORT" ]; then
  echo "Usage: $0 <port>"
  exit 1
fi

# Validate port
if ! [[ "$PORT" =~ ^[0-9]+$ ]] || [ "$PORT" -lt 1 ] || [ "$PORT" -gt 65535 ]; then
  echo "Error: Invalid port number: $PORT"
  exit 1
fi

# Kill the port
if tsunami "$PORT" -f; then
  echo "✓ Killed port $PORT"
else
  echo "✗ Failed to kill port $PORT"
  exit 1
fi
```

### Multi-Port Cleanup Script

```bash
#!/bin/bash
# cleanup-dev-ports.sh - Clean development ports

# Port configuration
readonly PORTS=(3000 8080 5432 6379)
readonly TIMEOUT="5s"

# Colors for output
readonly GREEN='\033[0;32m'
readonly RED='\033[0;31m'
readonly NC='\033[0m' # No Color

# Clean each port
echo "Cleaning development ports..."

FAILED=()

for PORT in "${PORTS[@]}"; do
  echo -n "Cleaning port $PORT... "

  if tsunami "$PORT" -f -q --timeout "$TIMEOUT"; then
    echo -e "${GREEN}✓${NC}"
  else
    echo -e "${RED}✗${NC}"
    FAILED+=("$PORT")
  fi
done

# Report results
echo ""
if [ ${#FAILED[@]} -eq 0 ]; then
  echo -e "${GREEN}All ports cleaned successfully${NC}"
  exit 0
else
  echo -e "${RED}Failed to clean ports: ${FAILED[*]}${NC}"
  exit 1
fi
```

## Advanced Script Patterns

### Port Manager Script

```bash
#!/bin/bash
# port-manager.sh - Comprehensive port management

set -euo pipefail

# Configuration
readonly SCRIPT_NAME=$(basename "$0")
readonly LOG_FILE="/tmp/port-manager.log"

# Logging
log() {
  echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

# Check if port is in use
is_port_used() {
  local port=$1
  tsunami -l --json | jq -e ".[] | select(.port == $port)" > /dev/null 2>&1
}

# Get process on port
get_process_on_port() {
  local port=$1
  tsunami -l --json | jq -r ".[] | select(.port == $port) | .process"
}

# Kill port with retry
kill_port_with_retry() {
  local port=$1
  local max_retries=3
  local retry=0

  while [ $retry -lt $max_retries ]; do
    if ! is_port_used "$port"; then
      log "Port $port is already free"
      return 0
    fi

    local process=$(get_process_on_port "$port")
    log "Attempting to kill $process on port $port (attempt $((retry + 1))/$max_retries)"

    if tsunami "$port" -f -q; then
      sleep 1
      if ! is_port_used "$port"; then
        log "Successfully killed port $port"
        return 0
      fi
    fi

    retry=$((retry + 1))
    sleep 2
  done

  log "Failed to kill port $port after $max_retries attempts"
  return 1
}

# Main logic
main() {
  local command=${1:-}
  shift || true

  case "$command" in
    kill)
      local port=$1
      kill_port_with_retry "$port"
      ;;

    list)
      log "Listing all ports"
      tsunami -l
      ;;

    check)
      local port=$1
      if is_port_used "$port"; then
        local process=$(get_process_on_port "$port")
        echo "Port $port is in use by: $process"
        exit 1
      else
        echo "Port $port is free"
        exit 0
      fi
      ;;

    clean)
      local ports=("$@")
      for port in "${ports[@]}"; do
        kill_port_with_retry "$port" || true
      done
      ;;

    *)
      echo "Usage: $SCRIPT_NAME <command> [args]"
      echo ""
      echo "Commands:"
      echo "  kill <port>           Kill process on port"
      echo "  list                  List all listening ports"
      echo "  check <port>          Check if port is in use"
      echo "  clean <port> [...]    Clean multiple ports"
      exit 1
      ;;
  esac
}

main "$@"
```

Usage:

```bash
./port-manager.sh kill 3000
./port-manager.sh list
./port-manager.sh check 8080
./port-manager.sh clean 3000 8080 5432
```

### Conditional Kill Script

```bash
#!/bin/bash
# conditional-kill.sh - Kill only if specific conditions are met

PORT=$1
PROCESS_FILTER=${2:-}

if [ -z "$PORT" ]; then
  echo "Usage: $0 <port> [process_filter]"
  exit 1
fi

# Get process info
PROCESS_JSON=$(tsunami -l --json | jq ".[] | select(.port == $PORT)")

if [ -z "$PROCESS_JSON" ]; then
  echo "No process on port $PORT"
  exit 0
fi

PROCESS_NAME=$(echo "$PROCESS_JSON" | jq -r '.process')
PROCESS_USER=$(echo "$PROCESS_JSON" | jq -r '.user')

# Apply filter if specified
if [ -n "$PROCESS_FILTER" ]; then
  if ! echo "$PROCESS_NAME" | grep -q "$PROCESS_FILTER"; then
    echo "Process '$PROCESS_NAME' doesn't match filter '$PROCESS_FILTER', skipping"
    exit 0
  fi
fi

# Check if it's our process
if [ "$PROCESS_USER" != "$USER" ]; then
  echo "Process is owned by $PROCESS_USER, not $USER. Use sudo?"
  exit 1
fi

# Kill it
echo "Killing $PROCESS_NAME (user: $PROCESS_USER) on port $PORT"
tsunami "$PORT" -f
```

## JSON Processing

### Parse and Kill Multiple Ports

```bash
#!/bin/bash
# kill-by-filter.sh - Kill all ports matching a filter

FILTER=$1

if [ -z "$FILTER" ]; then
  echo "Usage: $0 <process_filter>"
  exit 1
fi

# Get ports for processes matching filter
PORTS=$(tsunami -l --json | \
        jq -r ".[] | select(.process | contains(\"$FILTER\")) | .port" | \
        sort -u)

if [ -z "$PORTS" ]; then
  echo "No processes matching '$FILTER'"
  exit 0
fi

echo "Found processes matching '$FILTER' on ports:"
echo "$PORTS"
echo ""

read -p "Kill all? [y/N] " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
  echo "$PORTS" | xargs -I {} tsunami {} -f
  echo "✓ All matching processes killed"
else
  echo "Cancelled"
fi
```

### Generate Port Report

```bash
#!/bin/bash
# port-report.sh - Generate detailed port usage report

OUTPUT_FILE=${1:-port-report.txt}

{
  echo "Port Usage Report"
  echo "Generated: $(date)"
  echo "Hostname: $(hostname)"
  echo ""
  echo "===================================="
  echo ""

  # Get all ports
  tsunami -l --json | jq -r '.[] | "\(.port)|\(.pid)|\(.process)|\(.user)|\(.proto)"' | \
  while IFS='|' read -r port pid process user proto; do
    echo "Port: $port"
    echo "  PID: $pid"
    echo "  Process: $process"
    echo "  User: $user"
    echo "  Protocol: $proto"
    echo ""
  done

  echo "===================================="
  echo ""
  echo "Summary:"
  echo "  Total ports: $(tsunami -l --json | jq 'length')"
  echo "  Unique processes: $(tsunami -l --json | jq -r '.[].process' | sort -u | wc -l)"
  echo "  Unique users: $(tsunami -l --json | jq -r '.[].user' | sort -u | wc -l)"

} > "$OUTPUT_FILE"

echo "Report saved to: $OUTPUT_FILE"
cat "$OUTPUT_FILE"
```

## Loop and Monitoring Scripts

### Continuous Port Monitor

```bash
#!/bin/bash
# monitor-ports.sh - Continuously monitor and log port changes

INTERVAL=${1:-10}
LOG_FILE="/tmp/port-monitor.log"

echo "Monitoring ports every ${INTERVAL}s (log: $LOG_FILE)"
echo "Press Ctrl+C to stop"
echo ""

# Get initial state
LAST_STATE=$(tsunami -l --json | jq -r 'sort_by(.port) | .[]  | "\(.port):\(.pid)"')

while true; do
  # Get current state
  CURRENT_STATE=$(tsunami -l --json | jq -r 'sort_by(.port) | .[] | "\(.port):\(.pid)"')

  # Compare states
  if [ "$CURRENT_STATE" != "$LAST_STATE" ]; then
    TIMESTAMP=$(date +'%Y-%m-%d %H:%M:%S')

    # Log changes
    echo "[$TIMESTAMP] Port usage changed:" | tee -a "$LOG_FILE"

    # Find new ports
    NEW_PORTS=$(comm -13 <(echo "$LAST_STATE" | sort) <(echo "$CURRENT_STATE" | sort))
    if [ -n "$NEW_PORTS" ]; then
      echo "  New ports:" | tee -a "$LOG_FILE"
      echo "$NEW_PORTS" | while read -r line; do
        PORT=$(echo "$line" | cut -d: -f1)
        PROCESS=$(tsunami -l --json | jq -r ".[] | select(.port == $PORT) | .process")
        echo "    + Port $PORT: $PROCESS" | tee -a "$LOG_FILE"
      done
    fi

    # Find closed ports
    CLOSED_PORTS=$(comm -23 <(echo "$LAST_STATE" | sort) <(echo "$CURRENT_STATE" | sort))
    if [ -n "$CLOSED_PORTS" ]; then
      echo "  Closed ports:" | tee -a "$LOG_FILE"
      echo "$CLOSED_PORTS" | while read -r line; do
        PORT=$(echo "$line" | cut -d: -f1)
        echo "    - Port $PORT" | tee -a "$LOG_FILE"
      done
    fi

    echo "" | tee -a "$LOG_FILE"
    LAST_STATE=$CURRENT_STATE
  fi

  sleep "$INTERVAL"
done
```

### Auto-Restart on Port Conflict

```bash
#!/bin/bash
# auto-restart.sh - Auto-restart service on port conflict

PORT=$1
START_COMMAND=$2

if [ -z "$PORT" ] || [ -z "$START_COMMAND" ]; then
  echo "Usage: $0 <port> <start_command>"
  exit 1
fi

# Kill if port is in use
if tsunami -l --json | jq -e ".[] | select(.port == $PORT)" > /dev/null; then
  echo "Port $PORT is in use, killing..."
  tsunami "$PORT" -f
fi

# Start the service
echo "Starting: $START_COMMAND"
eval "$START_COMMAND" &
SERVICE_PID=$!

# Monitor for conflicts
while true; do
  sleep 5

  # Check if our service is still running
  if ! kill -0 "$SERVICE_PID" 2>/dev/null; then
    echo "Service died, exiting monitor"
    exit 1
  fi

  # Check for port conflicts
  CURRENT_PID=$(tsunami -l --json | jq -r ".[] | select(.port == $PORT) | .pid")

  if [ -n "$CURRENT_PID" ] && [ "$CURRENT_PID" != "$SERVICE_PID" ]; then
    echo "⚠️  Port conflict detected! Another process ($CURRENT_PID) is using port $PORT"
    echo "Killing conflicting process..."
    tsunami "$PORT" -f

    echo "Restarting our service..."
    kill "$SERVICE_PID" 2>/dev/null || true
    eval "$START_COMMAND" &
    SERVICE_PID=$!
  fi
done
```

## Integration Scripts

### Pre-Deployment Check

```bash
#!/bin/bash
# pre-deploy-check.sh - Check and free ports before deployment

set -e

REQUIRED_PORTS=(80 443 8080)
ENVIRONMENT=${1:-production}

echo "Pre-deployment check for $ENVIRONMENT"
echo "======================================"
echo ""

# Check each required port
CONFLICTS=()

for PORT in "${REQUIRED_PORTS[@]}"; do
  echo -n "Checking port $PORT... "

  if tsunami -l --json | jq -e ".[] | select(.port == $PORT)" > /dev/null; then
    PROCESS=$(tsunami -l --json | jq -r ".[] | select(.port == $PORT) | .process")
    echo "⚠️  IN USE by $PROCESS"
    CONFLICTS+=("$PORT:$PROCESS")
  else
    echo "✓ FREE"
  fi
done

# Handle conflicts
if [ ${#CONFLICTS[@]} -gt 0 ]; then
  echo ""
  echo "Port conflicts detected:"
  for conflict in "${CONFLICTS[@]}"; do
    echo "  - ${conflict}"
  done
  echo ""

  if [ "$ENVIRONMENT" = "production" ]; then
    # Production: ask for confirmation
    read -p "Kill conflicting processes? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      echo "Deployment cancelled"
      exit 1
    fi
  fi

  # Kill conflicts
  echo "Killing conflicting processes..."
  for PORT in "${REQUIRED_PORTS[@]}"; do
    tsunami "$PORT" -f --timeout 10s || true
  done

  echo "✓ Ports cleared"
fi

echo ""
echo "✓ Pre-deployment check passed"
```

### Health Check Script

```bash
#!/bin/bash
# health-check.sh - Check if services are running on expected ports

EXPECTED_PORTS=(3000 8080 5432)
EXIT_CODE=0

echo "Health Check"
echo "============"
echo ""

for PORT in "${EXPECTED_PORTS[@]}"; do
  echo -n "Port $PORT: "

  if PROCESS=$(tsunami -l --json | jq -r ".[] | select(.port == $PORT) | .process" 2>/dev/null) && [ -n "$PROCESS" ]; then
    echo "✓ UP ($PROCESS)"
  else
    echo "✗ DOWN"
    EXIT_CODE=1
  fi
done

echo ""
exit $EXIT_CODE
```

## Error Handling Patterns

### Robust Script Template

```bash
#!/bin/bash
# robust-killer.sh - Template for robust Tsunami scripts

set -euo pipefail  # Exit on error, undefined vars, pipe failures

# Configuration
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly LOG_FILE="/tmp/tsunami-script.log"

# Error handler
error_handler() {
  local line_num=$1
  echo "Error on line $line_num" >&2
  exit 1
}

trap 'error_handler ${LINENO}' ERR

# Logging
log() {
  local level=$1
  shift
  echo "[$(date +'%Y-%m-%d %H:%M:%S')] [$level] $*" | tee -a "$LOG_FILE"
}

# Validation
validate_port() {
  local port=$1

  if ! [[ "$port" =~ ^[0-9]+$ ]]; then
    log ERROR "Invalid port format: $port"
    return 1
  fi

  if [ "$port" -lt 1 ] || [ "$port" -gt 65535 ]; then
    log ERROR "Port out of range: $port"
    return 1
  fi

  return 0
}

# Main logic
main() {
  local port=$1

  if [ -z "$port" ]; then
    log ERROR "No port specified"
    echo "Usage: $0 <port>"
    exit 1
  fi

  if ! validate_port "$port"; then
    exit 1
  fi

  log INFO "Killing port $port"

  if tsunami "$port" -f; then
    log INFO "Successfully killed port $port"
  else
    log ERROR "Failed to kill port $port"
    exit 1
  fi
}

main "$@"
```

## Next Steps

- **[Development Workflows](./development-workflows.md)** - Real-world workflow examples
- **[Advanced Patterns](./advanced-patterns.md)** - Complex automation scenarios
- **[Best Practices](../guides/best-practices.md)** - Scripting best practices
