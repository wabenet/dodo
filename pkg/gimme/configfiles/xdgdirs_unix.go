// +build !windows

package configfiles

import (
	"os"
	"os/user"
	"path/filepath"
)

const (
	envXDGHome = "XDG_CONFIG_HOME"
	envXDGDirs = "XDG_CONFIG_DIRS"

	xdgDefaultDir = "/etc/etc/xdg"
)

func getXDGDirectories(name string) []string {
	var directories []string

	if xdgHome := os.Getenv(envXDGHome); xdgHome != "" {
		directories = append(directories, filepath.Join(xdgHome, name))
	} else if user, err := user.Current(); err == nil && user.HomeDir != "" {
		directories = append(directories, filepath.Join(user.HomeDir, ".config", name))
	}

	if xdgDirs := os.Getenv(envXDGDirs); xdgDirs != "" {
		for _, dir := range filepath.SplitList(xdgDirs) {
			directories = append(directories, filepath.Join(dir, name))
		}
	} else if xdgDefaultDir != "" {
		directories = append(directories, filepath.Join(xdgDefaultDir, name))
	}

	return directories
}
