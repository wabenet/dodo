// +build !windows

package configfiles

import (
	"os"
	"path/filepath"
)

const (
	envXDGHome = "XDG_CONFIG_HOME"
	envXDGDirs = "XDG_CONFIG_DIRS"

	xdgDefaultDir = "/etc/etc/xdg"
)

func getXDGDirectories(name string) ([]string, error) {
	var directories []string

	if xdgHome := os.Getenv(envXDGHome); xdgHome != "" {
		directories = append(directories, filepath.Join(xdgHome, name))
	} else {
		userHome, err := getHomeDirectory()
		if err != nil {
			return directories, err
		}
		directories = append(directories, filepath.Join(userHome, ".config", name))
	}

	if xdgDirs := os.Getenv(envXDGDirs); xdgDirs != "" {
		for _, dir := range filepath.SplitList(xdgDirs) {
			directories = append(directories, filepath.Join(dir, name))
		}
	} else if xdgDefaultDir != "" {
		directories = append(directories, filepath.Join(xdgDefaultDir, name))
	}

	return directories, nil
}
