package pull

import (
	"context"
	"strings"

	"github.com/fumiya-kume/mdefaults/internal/config"
	"github.com/fumiya-kume/mdefaults/internal/defaults"
)

func Pull(configs []config.Config) ([]config.Config, error) {
	defaultsCmds := make([]defaults.DefaultsCommand, 0, len(configs))
	for i := 0; i < len(configs); i++ {
		defaultsCmds = append(defaultsCmds, defaults.NewDefaultsCommandImpl(configs[i].Domain, configs[i].Key))
	}
	return PullImpl(defaultsCmds)
}

func PullImpl(defaultsCmds []defaults.DefaultsCommand) ([]config.Config, error) {
	updatedConfigs := make([]config.Config, 0, len(defaultsCmds))
	for i := 0; i < len(defaultsCmds); i++ {
		value, err := defaultsCmds[i].Read(context.Background())
		if err != nil {
			continue
		}
		value = strings.ReplaceAll(value, "\n", "")
		
		valueType, err := defaultsCmds[i].ReadType(context.Background())
		if err != nil {
			valueType = "string"
		}
		
		updatedConfigs = append(updatedConfigs, config.Config{
			Domain: defaultsCmds[i].Domain(),
			Key:    defaultsCmds[i].Key(),
			Value:  &value,
			Type:   valueType,
		})
	}
	return updatedConfigs, nil
}
