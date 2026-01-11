#!/bin/bash
# build-docs.sh - Build documentation using Volcano

set -e

echo "Building Tsunami documentation..."

# Check if volcano is installed
if ! command -v volcano &> /dev/null; then
  echo "Error: Volcano is not installed"
  echo "Install with: go install github.com/wusher/volcano@latest"
  exit 1
fi

# Build docs
volcano ./docs -o ./public \
  --title="Tsunami Documentation" \
  --url="https://wusher.github.io/tsunami" \
  --top-nav

echo "Documentation built successfully in ./public"
