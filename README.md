# mdefaults [![Go](https://github.com/fumiya-kume/mdefaults/actions/workflows/go.yml/badge.svg)](https://github.com/fumiya-kume/mdefaults/actions/workflows/go.yml)

mdefaults is a tool for Configuration as Code (CaC) for macOS. It allows you to manage macOS configuration settings through code, making it easier to version control and automate configuration changes.

https://github.com/user-attachments/assets/de6fe801-e8e1-400f-bbae-f3fdd9abb0a6

## How to Execute

### Building from Source

To build the program from source, run:

```bash
go build -o mdefaults ./cmd/mdefaults
```

This will create an executable file named `mdefaults` in the current directory.

### Running the Program

To run the program, use:

```bash
./mdefaults [command] [flags]
```

For example:

```bash
./mdefaults pull
./mdefaults push
./mdefaults --verbose pull
./mdefaults pull -y
```

See the [Usage](#usage) section below for more details on available commands and flags.

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

### pull

Pull the current macOS configuration that is written in the configuration file.

```
mdefaults pull
```

### push

Apply the configuration settings from the file to macOS.

```
mdefaults push
```


### Command Line Flags

#### Verbose Mode

Enable verbose logging to get detailed information about the application's operations. This can be useful for debugging and understanding the application's behavior.

To enable verbose mode, use the `--verbose` flag with any command:

```
mdefaults --verbose pull
```

This will provide additional log output in the console and write detailed logs to the `mdefaults.log` file.

#### Version Information

Print version and architecture information:

```
mdefaults --version
```

Or use the short form:

```
mdefaults -v
```

#### Auto-confirm

Automatically confirm prompts (useful for scripting):

```
mdefaults pull -y
```

### Debug Mode

Run the application in debug mode:

```
mdefaults debug
```

This command provides additional debugging information.

### Troubleshooting

If you encounter any issues, please check the following:
- Ensure the configuration file is correctly formatted.
- Verify that `mdefaults` is installed correctly.

### Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

### Testing

The project includes both unit tests and end-to-end (e2e) tests:

- **Unit Tests**: Run with `go test ./...`
- **E2E Tests**: Located in the `test/e2e` directory. These tests verify the tool's functionality in a real macOS environment. See the [E2E Tests README](test/e2e/README.md) for more information.

## Installation

```
go install github.com/fumiya-kume/mdefaults
```

## License

[GPL-3.0](LICENSE)
