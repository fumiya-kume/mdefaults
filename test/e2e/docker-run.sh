#!/bin/bash

# Exit on error
set -e

# Enable debug output
set -x

echo "Building and running mdefaults in Docker"

# Get the directory of the script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$SCRIPT_DIR/../.."

# Build the Docker image
echo "Building Docker image..."
docker build -t mdefaults -f "$SCRIPT_DIR/Dockerfile" "$REPO_ROOT"

# Run the Docker container with the provided arguments
echo "Running mdefaults in Docker..."
docker run --rm mdefaults "$@"

echo "Docker execution completed"