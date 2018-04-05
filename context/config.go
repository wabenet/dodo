package context

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

func (context *Context) ensureConfig() error {
	if context.Config != nil {
		return nil
	}
	if context.Options.Filename != "" {
		config, err := findConfigInFile(context.Name, context.Options.Filename)
		if err != nil {
			return err
		}
		context.Options.UpdateConfiguration(config)
		context.Config = config
		return nil
	}
	config, err := findConfigAnywhere(context.Name)
	if err != nil {
		return err
	}
	context.Options.UpdateConfiguration(config)
	context.Config = config
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

func findConfigAnywhere(contextName string) (*config.ContextConfig, error) {
	directories, err := findConfigDirectories()
	if err != nil {
		return nil, err
	}

	for _, directory := range directories {
		config, err := findConfigInDirectory(contextName, directory)
		if err == nil {
			return config, err
		}
		// TODO: log error
	}
	return nil, fmt.Errorf("Could not find configuration for context '%s' in any configuration file", contextName)
}

func findConfigInDirectory(contextName string, directory string) (*config.ContextConfig, error) {
	for _, filename := range configFileNames {
		path, _ := filepath.Abs(filepath.Join(directory, filename))
		// TODO: log error
		config, err := findConfigInFile(contextName, path)
		if err == nil {
			return config, err
		}
		// TODO: log error
	}
	return nil, fmt.Errorf("Could not find configuration for context '%s' in directory '%s'", directory)
}

// TODO: validation
// TODO: check if there are unknown keys
func findConfigInFile(contextName string, filename string) (*config.ContextConfig, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Could not read file %q", filename)
	}

	config := &config.Config{}
	err = yaml.Unmarshal(bytes, config)
	if err != nil {
		return nil, fmt.Errorf("Could not load config from %q: %s", filename, err)
	}

	if config.Contexts == nil {
		return nil, fmt.Errorf("File '%s' does not contain any context configurations", filename)
	}

	if contextConfig, ok := config.Contexts[contextName]; ok {
		return &contextConfig, nil
	}

	return nil, fmt.Errorf("File '%s' does not contain configuration for context '%s'", filename, contextName)
}
