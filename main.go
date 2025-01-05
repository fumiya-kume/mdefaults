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
		for i := 0; i < len(configs); i++ {
			pull(fs, configs, &DefaultsCommandImpl{
				domain: configs[i].Domain,
				key:    configs[i].Key,
			})
		}
	} else if command == "push" {
		push(configs)
	}
}

func pull(fs FileSystem, configs []Config, defaults DefaultsCommand) {
	updatedConfigs := make([]Config, 0, len(configs))
	for _, config := range configs {
		updatedConfig, err := readConfigFromMacos(config, defaults)
		if err != nil {
			fmt.Println(err)
			continue
		}
		updatedConfigs = append(updatedConfigs, updatedConfig)
	}
	writeConfigFile(fs, updatedConfigs)
}

func readConfigFromMacos(config Config, defaults DefaultsCommand) (Config, error) {
	value, err := defaults.Read(context.Background())
	if err != nil {
		return Config{}, err
	}
	config.Value = value
	return config, nil
}

func push(configs []Config) {
	for _, config := range configs {
		defaults := &DefaultsCommandImpl{
			domain: config.Domain,
			key:    config.Key,
		}
		defaults.Write(context.Background(), config.Value)
	}
}
