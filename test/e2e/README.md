# End-to-End Tests for mdefaults

This directory contains end-to-end tests for the mdefaults tool. These tests verify that the tool works correctly in a real macOS environment.

## Requirements

- macOS (the tests must be run on a macOS system)
- Go installed (to build the mdefaults binary if needed)

## Running the Tests

You can run the tests using the provided shell script:

```bash
./run_tests.sh
```

This script will:
1. Check if it's running on macOS
2. Set up a test environment
3. Back up any existing mdefaults configuration
4. Run a series of tests:
   - Creating a test configuration file
   - Running mdefaults pull
   - Modifying the configuration
   - Running mdefaults push
   - Verifying the changes were applied
   - Restoring the original values
5. Clean up after the tests

## Docker Support

A Dockerfile is provided to build the mdefaults binary for macOS. However, the tests themselves must be run on a macOS system, as Docker doesn't support running macOS containers.

To build the Docker image:

```bash
docker build -t mdefaults-e2e -f test/e2e/Dockerfile .
```

This will create a Docker image with the mdefaults binary built for macOS. You can then copy the binary from the Docker image to your macOS system:

```bash
docker create --name mdefaults-container mdefaults-e2e
docker cp mdefaults-container:/app/mdefaults .
docker rm mdefaults-container
```

Then run the tests on your macOS system:

```bash
./test/e2e/run_tests.sh
```

## GitHub Actions

These tests can also be run in GitHub Actions using a macOS runner. Here's an example workflow:

```yaml
name: E2E Tests

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  e2e-tests:
    runs-on: macos-latest
    steps:
    - uses: actions/checkout@v2
    - name: Set up Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.23
    - name: Build
      run: go build -o mdefaults ./cmd/mdefaults
    - name: Run E2E Tests
      run: ./test/e2e/run_tests.sh
```

This workflow will run the e2e tests on a macOS runner in GitHub Actions.