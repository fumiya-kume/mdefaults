# mdefaults [![Go](https://github.com/fumiya-kume/mdefaults/actions/workflows/go.yml/badge.svg)](https://github.com/fumiya-kume/mdefaults/actions/workflows/go.yml)

mdefaults is a tool for Configuration as Code (CaC) for macOS. It allows you to manage macOS configuration settings through code, making it easier to version control and automate configuration changes.

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

### Troubleshooting

If you encounter any issues, please check the following:
- Ensure the configuration file is correctly formatted.
- Verify that `mdefaults` is installed correctly.

### Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

## Installation

```
go install github.com/fumiya-kume/mdefaults
```

## License

[GPL-3.0](LICENSE)

