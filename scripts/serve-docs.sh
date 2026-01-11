#!/bin/bash
# serve-docs.sh - Serve documentation locally using Volcano

set -e

PORT=${1:-1776}

echo "Serving Tsunami documentation locally..."

# Check if volcano is installed
if ! command -v volcano &> /dev/null; then
  echo "Error: Volcano is not installed"
  echo "Install with: go install github.com/wusher/volcano@latest"
  exit 1
fi

echo ""
echo "Documentation available at: http://localhost:$PORT"
echo "Press Ctrl+C to stop the server"
echo ""

volcano -s -p "$PORT" ./docs
