#!/bin/bash
# serve-docs.sh - Serve documentation locally for preview

set -e

PORT=${1:-8000}

echo "Serving Tsunami documentation locally..."

# Check if volcano is installed
if ! command -v volcano &> /dev/null; then
  echo "Error: Volcano is not installed"
  echo "Install with: go install github.com/volcano/volcano/cmd/volcano@latest"
  exit 1
fi

# Check if docs are built
if [ ! -d "public" ]; then
  echo "Documentation not built yet. Building..."
  ./scripts/build-docs.sh
fi

# Serve docs
cd docs
volcano serve --port "$PORT"

echo ""
echo "Documentation available at: http://localhost:$PORT"
echo "Press Ctrl+C to stop the server"
