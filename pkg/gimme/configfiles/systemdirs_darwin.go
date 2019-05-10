// +build darwin

package configfiles

import (
	"os/user"
	"path/filepath"
)

func getUserDirectories(name string) []string {
	user, err := user.Current()
	if err != nil {
		return []string{}
	}
	if user.HomeDir == "" {
		return []string{}
	}
	return []string{
		user.HomeDir,
		filepath.Join(user.HomeDir, "Library", "Application Support"),
	}
}

func getSystemDirectories() []string {
	return []string{
		filepath.Join("/", "etc"),
		filepath.Join("/", "Library", "Application Support"),
	}
}
