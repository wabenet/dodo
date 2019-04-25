package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/oclaussen/dodo/pkg/configfiles"
	"github.com/oclaussen/dodo/pkg/types"
	"gopkg.in/yaml.v2"
)

// LoadConfiguration tries to find a backdrop configuration by name in any of
// the supported locations. If given, will only look in the supplied config
// file.
func LoadConfiguration(
	backdrop string, configfile string,
) (*types.Backdrop, error) {
	if configfile != "" {
		config, err := ParseConfigurationFile(configfile)
		if err != nil {
			return nil, err
		}
		result, ok := config.Backdrops[backdrop]
		if !ok {
			return nil, fmt.Errorf("could not find backdrop %s in file %s", backdrop, configfile)
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

	return nil, fmt.Errorf("could not find backdrop %s in any configuration file", backdrop)
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

// ParseConfigurationFile reads a full dodo configuration from a file.
func ParseConfigurationFile(filename string) (types.Group, error) {
	if !filepath.IsAbs(filename) {
		directory, err := os.Getwd()
		if err != nil {
			return types.Group{}, err
		}
		filename, err = filepath.Abs(filepath.Join(directory, filename))
		if err != nil {
			return types.Group{}, err
		}
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return types.Group{}, fmt.Errorf("could not read file '%s'", filename)
	}

	var mapType map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &mapType)
	if err != nil {
		return types.Group{}, err
	}

	config, err := types.DecodeGroup(filename, mapType)
	if err != nil {
		return types.Group{}, err
	}

	return config, nil
}
