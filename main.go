package main

import (
	"context"
	"flag"
	"fmt"
	"os"
)

func main() {
	command := os.Args[1]
	flag.Parse()
	fs := &fileSystem{}
	createConfigFileIfMissing(fs)
	configs, err := readConfigFile(fs)
	if err != nil {
		fmt.Println(err)
	}
	for i := 0; i < len(configs); i++ {
		fmt.Printf("- %s %s\n", configs[i].Domain, configs[i].Key)
	}

	switch command {
	case "pull":
		updatedConfigs, err := pull(configs)
		if err != nil {
			fmt.Println(err)
			return
		}
		writeConfigFile(fs, updatedConfigs)
	case "push":
		push(configs)
	default:
		fmt.Println("Invalid command")
	}
}

func pull(configs []Config) ([]Config, error) {
	defaults := make([]DefaultsCommand, 0, len(configs))
	for i := 0; i < len(configs); i++ {
		defaults = append(defaults, &DefaultsCommandImpl{
			domain: configs[i].Domain,
			key:    configs[i].Key,
		})
	}
	return pullImpl(defaults)
}

func pullImpl(defaults []DefaultsCommand) ([]Config, error) {
	updatedConfigs := make([]Config, 0, len(defaults))
	for i := 0; i < len(defaults); i++ {
		value, err := defaults[i].Read(context.Background())
		if err != nil {
			return nil, err
		}
		updatedConfigs = append(updatedConfigs, Config{
			Domain: defaults[i].Domain(),
			Key:    defaults[i].Key(),
			Value:  value,
		})
	}
	return updatedConfigs, nil
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
