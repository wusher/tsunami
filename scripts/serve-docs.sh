#!/bin/bash
# serve-docs.sh - Serve documentation locally for preview

set -e

PORT=${1:-1313}

echo "Serving Tsunami documentation locally..."

# Check if hugo is installed
if ! command -v hugo &> /dev/null; then
  echo "Error: Hugo is not installed"
  echo "Install with:"
  echo "  - macOS: brew install hugo"
  echo "  - Linux: sudo apt-get install hugo (or snap install hugo)"
  echo "  - Or download from: https://github.com/gohugoio/hugo/releases"
  exit 1
fi

# Serve docs (Hugo builds automatically in serve mode)
cd docs
echo ""
echo "Documentation available at: http://localhost:$PORT"
echo "Press Ctrl+C to stop the server"
echo ""
hugo server --port "$PORT" --bind 0.0.0.0
