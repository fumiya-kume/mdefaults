package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	command := os.Args[1]
	fs := &fileSystem{}
	createConfigFileIfMissing(fs)
	configs, err := readConfigFile(fs)
	if err != nil {
		fmt.Println(err)
	}
	for i := 0; i < len(configs); i++ {
		fmt.Printf("- %s %s\n", configs[i].Domain, configs[i].Key)
	}

	if command == "pull" {
		pull(fs, configs)
	} else if command == "push" {
		push(configs)
	}
}

func pull(fs FileSystem, configs []Config) {
	for _, config := range configs {
		readConfigFromMacos(config)
	}
	writeConfigFile(fs, configs)
}

func readConfigFromMacos(config Config) (Config, error) {
	defaults := DefaultsCommandImpl{
		domain: config.Domain,
		key:    config.Key,
	}

	value, err := defaults.Read(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	config.Value = value
	return config, nil
}

func push(configs []Config) {
	for _, config := range configs {
		defaults := DefaultsCommandImpl{
			domain: config.Domain,
			key:    config.Key,
		}
		defaults.Write(context.Background(), config.Value)
	}
}
