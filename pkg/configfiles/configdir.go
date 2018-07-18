package configfiles

import (
	"os"
	"os/user"
	"path/filepath"
)

var (
	xdgHome = os.Getenv("XDG_CONFIG_HOME")
	xdgDirs = os.Getenv("XDG_CONFIG_DIRS")
)

// FindConfigDirectories provides a list of directories on the file system
// that should be search for config files.
func FindConfigDirectories(appname string) ([]string, error) {
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

	if xdgHome != "" {
		configDir := filepath.Join(xdgHome, appname)
		configDirs = append(configDirs, configDir)
	} else {
		configdir := filepath.Join(user.HomeDir, ".config", appname)
		configDirs = append(configDirs, configdir)
	}

	if xdgDirs != "" {
		for _, xdgDir := range filepath.SplitList(xdgDirs) {
			configDir := filepath.Join(xdgDir, appname)
			configDirs = append(configDirs, configDir)
		}
	} else if xdgDefaultDir != "" {
		configDir := filepath.Join(xdgDefaultDir, appname)
		configDirs = append(configDirs, configDir)
	}

	for _, systemDir := range specialSystemDirectories {
		configDir := filepath.Join(systemDir, appname)
		configDirs = append(configDirs, configDir)
	}

	return configDirs, nil
}
