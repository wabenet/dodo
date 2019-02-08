package config

import (
	"fmt"

	"github.com/oclaussen/dodo/pkg/configfiles"
	"github.com/pkg/errors"
)

// LoadConfiguration tries to find a backdrop configuration by name in any of
// the supported locations. If given, will only look in the supplied config
// file.
func LoadConfiguration(
	backdrop string, configfile string,
) (*BackdropConfig, error) {
	if configfile != "" {
		config, err := ParseConfigurationFile(configfile)
		if err != nil {
			return nil, err
		}
		result, ok := config.Backdrops[backdrop]
		if !ok {
			return nil, errors.Errorf("Could not find backdrop %s in file %s", backdrop, configfile)
		}
		return &result, nil
	}

	candidates, err := configfiles.FindConfigFiles("dodo", []string{"yaml", "yml", "json"})
	if err != nil {
		return nil, err
	}

	for _, candidate := range candidates {
		config, err := ParseConfigurationFile(candidate)
		if err != nil {
			return nil, err
		}
		if result, ok := config.Backdrops[backdrop]; ok {
			return &result, nil
		}
	}

	return nil, errors.Errorf("Could not find backdrop %s in any configuration file", backdrop)
}

// ListConfigurations prints out all available backdrop names and the file
// it was found in.
func ListConfigurations() error {
	result := map[string]string{}
	candidates, err := configfiles.FindConfigFiles("dodo", []string{"yaml", "yml", "json"})
	if err != nil {
		return err
	}
	for _, candidate := range candidates {
		config, err := ParseConfigurationFile(candidate)
		if err != nil {
			return err
		}
		for name := range config.Backdrops {
			if result[name] == "" {
				fmt.Printf("%s (%s)\n", name, candidate)
				result[name] = candidate
			}
		}
	}
	return nil
}
