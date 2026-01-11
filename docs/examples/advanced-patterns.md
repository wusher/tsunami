# Advanced Patterns

Complex automation patterns and advanced Tsunami usage scenarios.

## Dynamic Port Discovery

### Find and Kill by Pattern

```bash
#!/bin/bash
# kill-by-pattern.sh - Find and kill processes matching complex patterns

# Kill all node processes on ports 3000-4000
tsunami -l --json | \
  jq -r '.[] | select(.process == "node" and .port >= 3000 and .port <= 4000) | .port' | \
  xargs -r -I {} tsunami {} -f

echo "✓ Killed all node processes in range 3000-4000"
```

### Kill by Age

```bash
#!/bin/bash
# kill-old-processes.sh - Kill processes running longer than threshold

MAX_AGE_SECONDS=${1:-3600}  # Default: 1 hour

echo "Killing processes older than $MAX_AGE_SECONDS seconds..."

tsunami -l --json | jq -r '.[].pid' | while read -r PID; do
  # Get process start time (seconds since epoch)
  START_TIME=$(ps -o lstart= -p "$PID" 2>/dev/null | xargs -I {} date -d "{}" +%s)

  if [ -z "$START_TIME" ]; then
    continue
  fi

  # Calculate age
  NOW=$(date +%s)
  AGE=$((NOW - START_TIME))

  if [ "$AGE" -gt "$MAX_AGE_SECONDS" ]; then
    # Get port for this PID
    PORT=$(tsunami -l --json | jq -r ".[] | select(.pid == $PID) | .port")
    PROCESS=$(tsunami -l --json | jq -r ".[] | select(.pid == $PID) | .process")

    echo "Killing old process: $PROCESS (PID $PID, port $PORT, age ${AGE}s)"
    tsunami "$PORT" -f || true
  fi
done
```

## Port Allocation Management

### Dynamic Port Allocator

```bash
#!/bin/bash
# allocate-port.sh - Find and allocate a free port in a range

START_PORT=${1:-3000}
END_PORT=${2:-3100}

# Get all used ports
USED_PORTS=$(tsunami -l --json | jq -r '.[].port' | sort -n)

# Find first free port
for ((PORT=START_PORT; PORT<=END_PORT; PORT++)); do
  if ! echo "$USED_PORTS" | grep -q "^${PORT}$"; then
    echo "$PORT"
    exit 0
  fi
done

echo "Error: No free ports in range $START_PORT-$END_PORT" >&2
exit 1
```

Usage:

```bash
# Find free port
FREE_PORT=$(./allocate-port.sh 3000 3100)

# Start service on that port
npm run dev -- --port "$FREE_PORT"
```

### Port Pool Manager

```bash
#!/bin/bash
# port-pool.sh - Manage a pool of reusable ports

POOL_FILE="/tmp/port-pool.txt"
MIN_PORT=3000
MAX_PORT=3100

# Initialize pool
init_pool() {
  : > "$POOL_FILE"
  for ((PORT=MIN_PORT; PORT<=MAX_PORT; PORT++)); do
    echo "$PORT:free" >> "$POOL_FILE"
  done
}

# Acquire port
acquire_port() {
  # Find first free port
  local port=$(grep ":free" "$POOL_FILE" | head -1 | cut -d: -f1)

  if [ -z "$port" ]; then
    echo "Error: No free ports available" >&2
    return 1
  fi

  # Make sure it's actually free
  if tsunami -l --json | jq -e ".[] | select(.port == $port)" > /dev/null; then
    # Port is in use, kill it
    tsunami "$port" -f -q || true
  fi

  # Mark as used
  sed -i "s/^${port}:free/${port}:used/" "$POOL_FILE"

  echo "$port"
}

# Release port
release_port() {
  local port=$1

  # Kill anything on the port
  tsunami "$port" -f -q || true

  # Mark as free
  sed -i "s/^${port}:used/${port}:free/" "$POOL_FILE"
}

# Main
case "${1:-}" in
  init)
    init_pool
    echo "Port pool initialized ($MIN_PORT-$MAX_PORT)"
    ;;
  acquire)
    acquire_port
    ;;
  release)
    release_port "$2"
    ;;
  status)
    echo "Port Pool Status:"
    echo "Free: $(grep -c ":free" "$POOL_FILE")"
    echo "Used: $(grep -c ":used" "$POOL_FILE")"
    ;;
  *)
    echo "Usage: $0 {init|acquire|release <port>|status}"
    exit 1
    ;;
esac
```

Usage:

```bash
# Initialize
./port-pool.sh init

# Acquire a port
PORT=$(./port-pool.sh acquire)
echo "Using port: $PORT"

# Start service
npm run dev -- --port "$PORT" &

# When done, release it
./port-pool.sh release "$PORT"
```

## Service Orchestration

### Service Manager

```bash
#!/bin/bash
# service-manager.sh - Manage multiple services with port tracking

SERVICES_DIR="/tmp/tsunami-services"
mkdir -p "$SERVICES_DIR"

# Start service
start_service() {
  local name=$1
  local port=$2
  local command=$3

  # Kill existing process on port
  tsunami "$port" -f -q || true

  # Start new service
  echo "Starting $name on port $port..."
  eval "$command" &
  local pid=$!

  # Save service info
  cat > "$SERVICES_DIR/$name.json" <<EOF
{
  "name": "$name",
  "port": $port,
  "pid": $pid,
  "command": "$command",
  "started": "$(date -Iseconds)"
}
EOF

  echo "✓ Started $name (PID: $pid, Port: $port)"
}

# Stop service
stop_service() {
  local name=$1
  local service_file="$SERVICES_DIR/$name.json"

  if [ ! -f "$service_file" ]; then
    echo "Service $name not found"
    return 1
  fi

  local port=$(jq -r '.port' "$service_file")
  local pid=$(jq -r '.pid' "$service_file")

  echo "Stopping $name (PID: $pid, Port: $port)..."

  # Kill by port
  tsunami "$port" -f || true

  # Also try killing by PID
  kill "$pid" 2>/dev/null || true

  # Remove service file
  rm "$service_file"

  echo "✓ Stopped $name"
}

# List services
list_services() {
  echo "Managed Services:"
  echo "================="

  if [ -z "$(ls -A "$SERVICES_DIR" 2>/dev/null)" ]; then
    echo "No services running"
    return
  fi

  for service_file in "$SERVICES_DIR"/*.json; do
    local name=$(jq -r '.name' "$service_file")
    local port=$(jq -r '.port' "$service_file")
    local pid=$(jq -r '.pid' "$service_file")
    local started=$(jq -r '.started' "$service_file")

    # Check if actually running
    if tsunami -l --json | jq -e ".[] | select(.port == $port)" > /dev/null; then
      local status="✓ Running"
    else
      local status="✗ Stopped"
    fi

    echo "$name: $status (PID: $pid, Port: $port, Started: $started)"
  done
}

# Restart service
restart_service() {
  local name=$1
  local service_file="$SERVICES_DIR/$name.json"

  if [ ! -f "$service_file" ]; then
    echo "Service $name not found"
    return 1
  fi

  local port=$(jq -r '.port' "$service_file")
  local command=$(jq -r '.command' "$service_file")

  stop_service "$name"
  sleep 1
  start_service "$name" "$port" "$command"
}

# Main
case "${1:-}" in
  start)
    start_service "$2" "$3" "$4"
    ;;
  stop)
    stop_service "$2"
    ;;
  restart)
    restart_service "$2"
    ;;
  list)
    list_services
    ;;
  clean)
    # Stop all services
    for service_file in "$SERVICES_DIR"/*.json; do
      name=$(jq -r '.name' "$service_file")
      stop_service "$name"
    done
    ;;
  *)
    echo "Usage: $0 {start|stop|restart|list|clean}"
    echo ""
    echo "Examples:"
    echo "  $0 start frontend 3000 'npm run dev'"
    echo "  $0 start backend 8080 'npm run api'"
    echo "  $0 list"
    echo "  $0 restart frontend"
    echo "  $0 stop backend"
    exit 1
    ;;
esac
```

Usage:

```bash
# Start services
./service-manager.sh start frontend 3000 'npm run dev'
./service-manager.sh start backend 8080 'npm run api'
./service-manager.sh start db 5432 'docker-compose up postgres'

# List all services
./service-manager.sh list

# Restart a service
./service-manager.sh restart frontend

# Stop a service
./service-manager.sh stop backend

# Stop all services
./service-manager.sh clean
```

## Load Balancer Port Management

### Round-Robin Port Killer

```bash
#!/bin/bash
# lb-rotate.sh - Kill one backend instance at a time for rolling restart

BACKEND_PORTS=(8081 8082 8083 8084)
RESTART_COMMAND="npm run backend"
HEALTH_CHECK_URL="http://localhost"

for PORT in "${BACKEND_PORTS[@]}"; do
  echo "Rotating backend on port $PORT..."

  # Kill this instance
  tsunami "$PORT" -f

  # Start new instance
  $RESTART_COMMAND --port "$PORT" &
  NEW_PID=$!

  # Wait for health check
  echo "Waiting for port $PORT to be healthy..."
  for i in {1..30}; do
    if curl -sf "${HEALTH_CHECK_URL}:${PORT}/health" > /dev/null; then
      echo "✓ Port $PORT is healthy"
      break
    fi
    sleep 1
  done

  # Wait a bit before rotating next
  sleep 5
done

echo "✓ All backends rotated"
```

## Conflict Resolution

### Multi-User Port Resolver

```bash
#!/bin/bash
# resolve-conflicts.sh - Resolve port conflicts on shared dev machine

# Configuration
readonly RESERVED_PORTS=(80 443 22 3306 5432 6379)

echo "Scanning for port conflicts..."

# Get all ports grouped by port number
tsunami -l --json | jq -r 'group_by(.port) | .[] | select(length > 1) | .[0].port' | \
while read -r PORT; do
  echo ""
  echo "Conflict on port $PORT:"

  # Show all processes on this port
  tsunami -l --json | jq -r ".[] | select(.port == $PORT) | \"  PID \(.pid): \(.process) (user: \(.user))\""

  # Check if it's a reserved port
  if [[ " ${RESERVED_PORTS[*]} " =~ " ${PORT} " ]]; then
    echo "  ⚠️  This is a reserved system port!"
  fi

  # Ask what to do
  echo ""
  echo "Options:"
  echo "  1) Kill all"
  echo "  2) Keep first, kill others"
  echo "  3) Skip"
  read -p "Choose [1-3]: " -n 1 -r
  echo ""

  case $REPLY in
    1)
      tsunami "$PORT" --all -f
      echo "✓ Killed all processes on port $PORT"
      ;;
    2)
      # Get all PIDs except first
      PIDS=$(tsunami -l --json | jq -r ".[] | select(.port == $PORT) | .pid" | tail -n +2)
      for PID in $PIDS; do
        tsunami --pid "$PID" -f
      done
      echo "✓ Kept first process, killed others"
      ;;
    3)
      echo "Skipped port $PORT"
      ;;
  esac
done

echo ""
echo "✓ Conflict resolution complete"
```

## Testing Infrastructure

### Test Port Isolation

```bash
#!/bin/bash
# test-isolation.sh - Ensure test port isolation

TEST_ID=${1:-$(date +%s)}
BASE_PORT=3000
PORTS_PER_TEST=10

# Calculate port range for this test
START_PORT=$((BASE_PORT + (TEST_ID % 100) * PORTS_PER_TEST))
END_PORT=$((START_PORT + PORTS_PER_TEST - 1))

echo "Test isolation for test ID: $TEST_ID"
echo "Port range: $START_PORT-$END_PORT"

# Clean our range
tsunami "$START_PORT-$END_PORT" -f -q || true

# Export for test runner
export TEST_PORT_START=$START_PORT
export TEST_PORT_END=$END_PORT

# Run tests
npm test

# Cleanup
tsunami "$START_PORT-$END_PORT" -f -q || true
```

### Parallel Test Runner with Port Pools

```bash
#!/bin/bash
# parallel-test-runner.sh - Run tests in parallel with isolated ports

NUM_WORKERS=${1:-4}
BASE_PORT=3000
PORTS_PER_WORKER=100

# Cleanup function
cleanup() {
  echo "Cleaning up all test ports..."
  tsunami "$BASE_PORT-$((BASE_PORT + NUM_WORKERS * PORTS_PER_WORKER))" -f -q || true
}

trap cleanup EXIT

# Start workers
for ((WORKER=0; WORKER<NUM_WORKERS; WORKER++)); do
  WORKER_START=$((BASE_PORT + WORKER * PORTS_PER_WORKER))
  WORKER_END=$((WORKER_START + PORTS_PER_WORKER - 1))

  echo "Starting worker $WORKER (ports $WORKER_START-$WORKER_END)..."

  (
    # Clean worker ports
    tsunami "$WORKER_START-$WORKER_END" -f -q || true

    # Run tests for this worker
    TEST_PORT_START=$WORKER_START TEST_PORT_END=$WORKER_END npm test

    # Cleanup worker ports
    tsunami "$WORKER_START-$WORKER_END" -f -q || true
  ) &
done

# Wait for all workers
wait

echo "✓ All workers complete"
```

## Monitoring and Alerting

### Port Usage Alert System

```bash
#!/bin/bash
# port-alert.sh - Alert when ports exceed threshold

THRESHOLD=50
ALERT_EMAIL="devops@example.com"

# Count listening ports
PORT_COUNT=$(tsunami -l --json | jq 'length')

if [ "$PORT_COUNT" -gt "$THRESHOLD" ]; then
  echo "⚠️  ALERT: High port usage detected ($PORT_COUNT ports)"

  # Generate report
  REPORT="/tmp/port-alert-$(date +%s).txt"
  {
    echo "Port Usage Alert"
    echo "================"
    echo "Time: $(date)"
    echo "Total ports: $PORT_COUNT"
    echo "Threshold: $THRESHOLD"
    echo ""
    echo "Top processes:"
    tsunami -l --json | jq -r '.[].process' | sort | uniq -c | sort -rn | head -10
    echo ""
    echo "Full port list:"
    tsunami -l
  } > "$REPORT"

  # Send alert (example with mail)
  # mail -s "Port Usage Alert: $PORT_COUNT ports" "$ALERT_EMAIL" < "$REPORT"

  cat "$REPORT"
fi
```

## Next Steps

- **[Development Workflows](./development-workflows.md)** - Practical workflow examples
- **[Scripting & Automation](./scripting.md)** - Script automation patterns
- **[Best Practices](../guides/best-practices.md)** - Tips for effective usage
