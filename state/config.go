package state

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/oclaussen/dodo/config"
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

func (state *state) ensureConfig() error {
	if state.Config != nil {
		return nil
	}
	if state.Options.Filename != "" {
		config, err := findConfigInFile(state.Name, state.Options.Filename)
		if err != nil {
			return err
		}
		state.Options.UpdateConfiguration(config)
		state.Config = config
		return nil
	}
	config, err := findConfigAnywhere(state.Name)
	if err != nil {
		return err
	}
	state.Options.UpdateConfiguration(config)
	state.Config = config
	return nil
}

func findConfigDirectories() ([]string, error) {
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

func findConfigAnywhere(backdrop string) (*config.BackdropConfig, error) {
	directories, err := findConfigDirectories()
	if err != nil {
		return nil, err
	}

	for _, directory := range directories {
		config, err := findConfigInDirectory(backdrop, directory)
		if err == nil {
			return config, err
		}
		// TODO: log error
	}
	return nil, fmt.Errorf("Could not find configuration for backdrop '%s' in any configuration file", backdrop)
}

func findConfigInDirectory(backdrop string, directory string) (*config.BackdropConfig, error) {
	for _, filename := range configFileNames {
		path, _ := filepath.Abs(filepath.Join(directory, filename))
		// TODO: log error
		config, err := findConfigInFile(backdrop, path)
		if err == nil {
			return config, err
		}
		// TODO: log error
	}
	return nil, fmt.Errorf("Could not find configuration for backdrop '%s' in directory '%s'", directory)
}

// TODO: validation
// TODO: check if there are unknown keys
func findConfigInFile(backdrop string, filename string) (*config.BackdropConfig, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Could not read file %q", filename)
	}

	config := &config.Config{}
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
