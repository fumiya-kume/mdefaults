#!/bin/bash

# Exit on error
set -e

# Enable debug output
set -x

echo "Starting mdefaults e2e tests"

# Check if running on macOS
if [[ "$(uname)" != "Darwin" ]]; then
    echo "Error: This test script must be run on macOS"
    exit 1
fi

# Set up test environment
TEST_DIR=$(mktemp -d)
CONFIG_DIR="$HOME/.mdefaults"
BACKUP_DIR="$HOME/.mdefaults.backup"

# Backup existing configuration if it exists
if [ -d "$CONFIG_DIR" ]; then
    echo "Backing up existing configuration"
    mv "$CONFIG_DIR" "$BACKUP_DIR"
fi

# Create test configuration directory
mkdir -p "$CONFIG_DIR"

# Function to clean up after tests
cleanup() {
    echo "Cleaning up test environment"
    rm -rf "$TEST_DIR"
    rm -f "$CONFIG_DIR/config"

    # Restore backup if it exists
    if [ -d "$BACKUP_DIR" ]; then
        rm -rf "$CONFIG_DIR"
        mv "$BACKUP_DIR" "$CONFIG_DIR"
    fi

    echo "Cleanup complete"
}

# Register cleanup function to run on exit
trap cleanup EXIT

# Get the directory of the script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Path to the mdefaults binary
MDEFAULTS_BIN="$SCRIPT_DIR/../../mdefaults"

# If the binary doesn't exist, try to build it
if [ ! -f "$MDEFAULTS_BIN" ]; then
    echo "mdefaults binary not found, building it"
    cd "$SCRIPT_DIR/../.."
    go build -o mdefaults ./cmd/mdefaults
    MDEFAULTS_BIN="$SCRIPT_DIR/../../mdefaults"
fi

# Test 1: Create a test configuration file
echo "Test 1: Creating test configuration file"
cat > "$CONFIG_DIR/config" << EOF
com.apple.dock autohide
com.apple.finder ShowPathbar
EOF

# Test 2: Run mdefaults pull
echo "Test 2: Running mdefaults pull"
"$MDEFAULTS_BIN" pull --yes

# Verify the configuration file was updated
if [ ! -f "$CONFIG_DIR/config" ]; then
    echo "Error: Configuration file not found after pull"
    exit 1
fi

# Test 3: Modify the configuration file
echo "Test 3: Modifying configuration file"
# Save the original value of autohide
ORIGINAL_AUTOHIDE=$(defaults read com.apple.dock autohide)
# Toggle the value
if [ "$ORIGINAL_AUTOHIDE" = "1" ]; then
    NEW_AUTOHIDE="false"
else
    NEW_AUTOHIDE="true"
fi

# Update the configuration file
cat > "$CONFIG_DIR/config" << EOF
com.apple.dock autohide $NEW_AUTOHIDE
com.apple.finder ShowPathbar
EOF

# Test 4: Run mdefaults push
echo "Test 4: Running mdefaults push"
"$MDEFAULTS_BIN" push

# Verify the changes were applied
CURRENT_AUTOHIDE=$(defaults read com.apple.dock autohide)
EXPECTED_VALUE="$NEW_AUTOHIDE"
if [ "$NEW_AUTOHIDE" = "true" ] && [ "$CURRENT_AUTOHIDE" = "1" ]; then
    # Values match (true is represented as 1)
    echo "Configuration applied correctly"
elif [ "$NEW_AUTOHIDE" = "false" ] && [ "$CURRENT_AUTOHIDE" = "0" ]; then
    # Values match (false is represented as 0)
    echo "Configuration applied correctly"
else
    echo "Error: Configuration not applied correctly"
    echo "Expected: $NEW_AUTOHIDE, Got: $CURRENT_AUTOHIDE"
    exit 1
fi

# Test 5: Restore original value
echo "Test 5: Restoring original value"
cat > "$CONFIG_DIR/config" << EOF
com.apple.dock autohide $ORIGINAL_AUTOHIDE
com.apple.finder ShowPathbar
EOF

"$MDEFAULTS_BIN" push

# Verify the original value was restored
CURRENT_AUTOHIDE=$(defaults read com.apple.dock autohide)
if [ "$CURRENT_AUTOHIDE" = "$ORIGINAL_AUTOHIDE" ]; then
    echo "Original value restored correctly"
else
    echo "Error: Original value not restored correctly"
    echo "Expected: $ORIGINAL_AUTOHIDE, Got: $CURRENT_AUTOHIDE"
    exit 1
fi

echo "All tests passed successfully!"
exit 0
