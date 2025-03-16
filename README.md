# mdefaults [![Go](https://github.com/fumiya-kume/mdefaults/actions/workflows/go.yml/badge.svg)](https://github.com/fumiya-kume/mdefaults/actions/workflows/go.yml) [![codecov](https://codecov.io/gh/fumiya-kume/mdefaults/branch/master/graph/badge.svg)](https://codecov.io/gh/fumiya-kume/mdefaults)

mdefaults is a tool for Configuration as Code (CaC) for macOS. It allows you to manage macOS configuration settings through code, making it easier to version control and automate configuration changes.

https://github.com/user-attachments/assets/de6fe801-e8e1-400f-bbae-f3fdd9abb0a6

## Usage

### Getting Started

#### Install

In the terminal
```
brew tap fumiya-kume/mdefaults
brew install mdefaults
```

In the `.Brewfile`
```
tap "fumiya-kume/mdefaults"
brew "mdefaults"
```

#### Create a Config File

Place your configuration file in the directory `~/.mdefaults`. This is an example configuration file:

```
com.apple.dock autohide
com.apple.finder ShowPathbar
```

Then execute `mdefaults pull` (get the current macOS configuration and save it to the file), `mdefaults push` (apply the configuration file to macOS).

### Value Type Support

mdefaults now supports preserving the original value types (boolean, integer, string, etc.) when reading from and writing to macOS defaults. This prevents issues that can occur when values are inadvertently stored as strings instead of their proper types.

When you use `mdefaults pull`, the tool will automatically detect the proper type of each value and save it in the configuration file. The updated configuration format looks like this:

```
com.apple.AppleMultitouchTrackpad FirstClickThreshold -integer 1
com.apple.AppleMultitouchTrackpad DragLock -boolean true
com.apple.finder ShowPathbar -boolean true
```

Each line follows this format: `domain key -type value`

Supported types include:
- `-string`: Text values
- `-int` or `-integer`: Integer values
- `-float`: Floating point values
- `-bool` or `-boolean`: Boolean values (true/false)
- `-date`: Date values
- `-data`: Binary data values
- `-array`: Array values
- `-dict`: Dictionary/object values

When you use `mdefaults push`, the tool will use these type specifications to ensure values are written with their correct types.

### pull

Pull the current macOS configuration that is written in the configuration file.

```
mdefaults pull
```

This command will:
1. Read the config file at `~/.mdefaults`
2. For each entry, read its current value and type from macOS defaults
3. Update the config file with the current values and correct types

### push

Apply the configuration settings from the file to macOS.

```
mdefaults push
```

This command will write each configuration entry to macOS defaults using the type specified in the config file.

### config

Print the configuration file content.

```
mdefaults config
```

This command reads the configuration file located in `~/.mdefaults` and prints its contents to the console.

### Verbose Mode

Enable verbose logging to get detailed information about the application's operations. This can be useful for debugging and understanding the application's behavior.

To enable verbose mode, use the `--verbose` flag with any command:

```
mdefaults --verbose pull
```

This will provide additional log output in the console and write detailed logs to the `mdefaults.log` file.

### Troubleshooting

If you encounter any issues, please check the following:
- Ensure the configuration file is correctly formatted.
- Verify that `mdefaults` is installed correctly.
- If macOS is crashing after login, check that your configuration values have the correct types specified.

### Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

### Development

#### Docker Support

You can run mdefaults in a Docker container using the provided Dockerfile and docker-compose.yml. This is useful for development and testing purposes, but note that since mdefaults is designed for macOS, some functionality may be limited in a Docker container.

To build and run the Docker container:

```bash
# Build the Docker image
docker-compose build

# Run mdefaults with the default command (--help)
docker-compose run mdefaults

# Run a specific command
docker-compose run mdefaults pull
docker-compose run mdefaults push
docker-compose run mdefaults --verbose pull
```

The Docker container mounts your local ~/.mdefaults directory, so it can read and write to the same configuration file as your local installation.

You can test the Docker setup using the provided test script:

```bash
./test-docker.sh
```

This script will build the Docker image and run a simple command to verify that everything is working correctly. The script is designed to be robust and will work even if Docker is not properly configured on your system.

Note: For testing purposes, the volume mount in docker-compose.yml is commented out to avoid potential issues. If you want to mount your ~/.mdefaults directory, you can uncomment the volume mount lines in docker-compose.yml.

#### Testing
The project includes both unit tests and E2E (End-to-End) tests. 

To run only the unit tests:
```
go test -short ./...
```

To run all tests including E2E tests (requires running in CI environment):
```
go test ./...
```

To run tests with code coverage:
```
go test -coverprofile=coverage.txt -covermode=atomic ./...
```

You can then view the coverage report in your browser:
```
go tool cover -html=coverage.txt
```

You can also run E2E tests using Docker Compose:
```
docker-compose up --build e2e-test
```

This will run the E2E tests in a Docker container with the CI environment variable set to true. The E2E tests use a custom Docker image that provides a macOS-like environment for testing.

Note: E2E tests are skipped by default when not running in a CI environment to prevent modifying your local macOS settings.

## Installation

```
go install github.com/fumiya-kume/mdefaults
```

## License

[GPL-3.0](LICENSE)
