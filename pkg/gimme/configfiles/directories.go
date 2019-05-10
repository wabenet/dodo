package configfiles

import (
	"os"
	"path/filepath"
)

func GimmeConfigDirectories(opts *Options) ([]string, error) {
	var configDirs []string

	if err := opts.normalize(); err != nil {
		return configDirs, err
	}

	if opts.IncludeWorkingDirectories {
		configDirs = append(configDirs, getWorkingDirectories()...)
	}

	configDirs = append(configDirs, getUserDirectories(opts.Name)...)
	configDirs = append(configDirs, getXDGDirectories(opts.Name)...)
	configDirs = append(configDirs, getSystemDirectories()...)
	return uniqueStrings(configDirs), nil
}

func getWorkingDirectories() []string {
	workingDir, err := os.Getwd()
	if err != nil {
		return []string{}
	}

	directories := []string{}
	for directory := workingDir; !isFSRoot(directory); directory = filepath.Dir(directory) {
		directories = append(directories, directory)
	}

	return directories
}
