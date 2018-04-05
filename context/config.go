package context

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/oclaussen/dodo/config"
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
		context.Config = context.applyOptionsToConfig(config)
		return nil
	}
	config, err := findConfigAnywhere(context.Name)
	if err != nil {
		return err
	}
	context.Config = context.applyOptionsToConfig(config)
	return nil
}

// TODO: this is a weird location for this
func (context *Context) applyOptionsToConfig(config *config.ContextConfig) *config.ContextConfig {
	if context.Options.Pull {
		config.Pull = true
	}
	if context.Options.Remove {
		remove := true
		config.Remove = &remove
	}
	if config.Build != nil {
		if context.Options.NoCache {
			config.Build.NoCache = true
		}
	}
	return config
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

func findConfigInFile(contextName string, filename string) (*config.ContextConfig, error) {
	config, err := config.Load(filename)
	if err != nil {
		return nil, err
	}
	if config.Contexts == nil {
		return nil, fmt.Errorf("File '%s' does not contain any context configurations", filename)
	}
	if contextConfig, ok := config.Contexts[contextName]; ok {
		return &contextConfig, nil
	}
	return nil, fmt.Errorf("File '%s' does not contain configuration for context '%s'", filename, contextName)
}
