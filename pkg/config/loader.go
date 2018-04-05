package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var (
	configFileNames = []string{
		"dodo.yaml",
		"dodo.yml",
		"dodo.json",
		".dodo.yaml",
		".dodo.yml",
		".dodo.json",
	}
)

func LoadConfiguration(backdrop string, configfile string) (*BackdropConfig, error) {
	if configfile == "" {
		return FindConfigAnywhere(backdrop)
	} else {
		return FindConfigInFile(backdrop, configfile)
	}
}

func FindConfigDirectories() ([]string, error) {
	var configDirectories []string

	workingDir, err := os.Getwd()
	if err != nil {
		return configDirectories, err
	}
	for directory := workingDir; directory != "/"; directory = filepath.Dir(directory) {
		configDirectories = append(configDirectories, directory)
	}
	configDirectories = append(configDirectories, "/")

	user, err := user.Current()
	if err != nil {
		return configDirectories, err
	}
	configDirectories = append(configDirectories, user.HomeDir)
	configDirectories = append(configDirectories, filepath.Join(user.HomeDir, ".config", "dodo"))

	configDirectories = append(configDirectories, "/etc")

	return configDirectories, nil
}

func FindConfigAnywhere(backdrop string) (*BackdropConfig, error) {
	directories, err := FindConfigDirectories()
	if err != nil {
		return nil, err
	}

	for _, directory := range directories {
		config, err := FindConfigInDirectory(backdrop, directory)
		if err == nil {
			return config, err
		}
		// TODO: log error
	}
	return nil, fmt.Errorf("Could not find configuration for backdrop '%s' in any configuration file", backdrop)
}

func FindConfigInDirectory(backdrop string, directory string) (*BackdropConfig, error) {
	for _, filename := range configFileNames {
		path, _ := filepath.Abs(filepath.Join(directory, filename))
		// TODO: log error
		config, err := FindConfigInFile(backdrop, path)
		if err == nil {
			return config, err
		}
		// TODO: log error
	}
	return nil, fmt.Errorf("Could not find configuration for backdrop '%s' in directory '%s'", directory)
}

// TODO: validation
// TODO: check if there are unknown keys
func FindConfigInFile(backdrop string, filename string) (*BackdropConfig, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Could not read file %q", filename)
	}

	config := &Config{}
	err = yaml.Unmarshal(bytes, config)
	if err != nil {
		return nil, fmt.Errorf("Could not load config from %q: %s", filename, err)
	}

	if config.Backdrops == nil {
		return nil, fmt.Errorf("File '%s' does not contain any backdrop configurations", filename)
	}

	if backdropConfig, ok := config.Backdrops[backdrop]; ok {
		return &backdropConfig, nil
	}

	return nil, fmt.Errorf("File '%s' does not contain configuration for backdrop '%s'", filename, backdrop)
}
