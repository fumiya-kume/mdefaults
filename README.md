# mdefaults [![Go](https://github.com/fumiya-kume/mdefaults/actions/workflows/go.yml/badge.svg)](https://github.com/fumiya-kume/mdefaults/actions/workflows/go.yml)

mdefaults is an tool for Configuration as Code (CaC) for macOS. 

## Usage

### Getting started

#### get the source code and execute it

```
go install github.com/fumiya-kume/mdefaults
mdefaults pull
```

#### Create a config file (or tool will create empty configuration file)

Place your configuration file in the directory of `~/.mdefaults`.
This is example condiguration file

```
com.apple.dock autohide
com.apple.finder ShowPathbar
``` 

and thee execute `mdefaults pull`(get the current macOS configuration and save it to the file), `mdefaults push`(apply the configuration file to the macOS).

### pull

Pull the current macOS configuration which wrote in the configuration file.

```
mdefaults pull
```

### push

Push the configuration file to the macOS.

```
mdefaults push
```


## Installation

```
go install github.com/fumiya-kume/mdefaults
```

## License

[GPL-3.0](LICENSE)

