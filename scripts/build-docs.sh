#!/bin/bash
# build-docs.sh - Build documentation locally with Hugo

set -e

echo "Building Tsunami documentation with Hugo..."

# Check if hugo is installed
if ! command -v hugo &> /dev/null; then
  echo "Error: Hugo is not installed"
  echo "Install with:"
  echo "  - macOS: brew install hugo"
  echo "  - Linux: sudo apt-get install hugo (or snap install hugo)"
  echo "  - Or download from: https://github.com/gohugoio/hugo/releases"
  exit 1
fi

# Build docs
cd docs
hugo --minify --destination ../public

echo ""
echo "✓ Documentation built successfully!"
echo ""
echo "Output directory: $(pwd)/../public"
echo "To serve locally, run: ./scripts/serve-docs.sh"
