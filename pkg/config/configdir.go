package config

import (
	"os"
	"os/user"
	"path/filepath"
)

const (
	XDGHome = "XDG_CONFIG_HOME"
	XDGDirs = "XDG_CONFIG_DIRS"
	DirName = "dodo"
)

// FindConfigDirectories provides a list of directories on the file system
// that should be search for config files.
func FindConfigDirectories() ([]string, error) {
	var configDirs []string

	workingDir, err := os.Getwd()
	if err != nil {
		return configDirs, err
	}
	user, err := user.Current()
	if err != nil {
		return configDirs, err
	}

	for directory := workingDir; directory != "/"; directory = filepath.Dir(directory) {
		configDirs = append(configDirs, directory)
	}

	if user.HomeDir != "" {
		configDirs = append(configDirs, user.HomeDir)
	}

	if xdgHome := os.Getenv(XDGHome); xdgHome != "" {
		configDir := filepath.Join(xdgHome, DirName)
		configDirs = append(configDirs, configDir)
	} else {
		configdir := filepath.Join(user.HomeDir, ".config", DirName)
		configDirs = append(configDirs, configdir)
	}

	if xdgDirs := os.Getenv(XDGDirs); xdgDirs != "" {
		for _, xdgDir := range filepath.SplitList(xdgDirs) {
			configDir := filepath.Join(xdgDir, DirName)
			configDirs = append(configDirs, configDir)
		}
	} else if XDGDefaultDir != "" {
		configDir := filepath.Join(XDGDefaultDir, DirName)
		configDirs = append(configDirs, configDir)
	}

	for _, systemDir := range SpecialSystemDirectories {
		configDir := filepath.Join(systemDir, DirName)
		configDirs = append(configDirs, configDir)
	}

	return configDirs, nil
}
