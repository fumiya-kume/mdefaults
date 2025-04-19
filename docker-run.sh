#!/usr/bin/env bash
set -euo pipefail

# This script builds and runs the mdefaults Docker image.
# Usage: ./docker-run.sh [mdefaults-args]

# Determine script directory (project root)
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

IMAGE_NAME="mdefaults"

echo "Building Docker image '${IMAGE_NAME}'..."
docker build -t "${IMAGE_NAME}" "${SCRIPT_DIR}"

echo "Running mdefaults in Docker..."
docker run --rm \
  -v "$HOME/.mdefaults":"$HOME/.mdefaults" \
  -e HOME="$HOME" \
  "${IMAGE_NAME}" "$@"