#!/bin/bash
# build-docs.sh - Build documentation locally with Volcano

set -e

echo "Building Tsunami documentation with Volcano..."

# Check if volcano is installed
if ! command -v volcano &> /dev/null; then
  echo "Error: Volcano is not installed"
  echo "Install with: go install github.com/wusher/volcano@latest"
  exit 1
fi

# Build docs
cd docs
volcano build . --output ../public --url https://wusher.github.io/tsunami/

echo ""
echo "✓ Documentation built successfully!"
echo ""
echo "Output directory: $(pwd)/../public"
