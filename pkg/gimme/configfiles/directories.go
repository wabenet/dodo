package configfiles

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
)

func GimmeConfigDirectories(opts *Options) ([]string, error) {
	var configDirs []string

	if err := opts.normalize(); err != nil {
		return configDirs, err
	}

	if opts.IncludeWorkingDirectories {
		directories, err := getWorkingDirectories()
		if err != nil {
			return configDirs, err
		}
		configDirs = append(configDirs, directories...)
	}

	homeDir, err := getHomeDirectory()
	if err != nil {
		return configDirs, err
	}
	configDirs = append(configDirs, homeDir)

	xdgDirs, err := getXDGDirectories(opts.Name)
	if err != nil {
		return configDirs, err
	}
	configDirs = append(configDirs, xdgDirs...)

	systemDirs, err := getSystemDirectories(opts.Name)
	if err != nil {
		return configDirs, err
	}
	configDirs = append(configDirs, systemDirs...)

	return configDirs, nil
}

func getHomeDirectory() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	if user.HomeDir == "" {
		return "", errors.New("current user has no home directory")
	}
	return user.HomeDir, nil
}

func getWorkingDirectories() ([]string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return []string{}, err
	}

	// TODO: check if filepath.Clean(path) ends in separator
	directories := []string{}
	for directory := workingDir; directory != "/"; directory = filepath.Dir(directory) {
		directories = append(directories, directory)
	}

	return directories, nil
}
