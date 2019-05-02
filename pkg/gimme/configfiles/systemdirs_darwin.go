// +build darwin

package configfiles

import (
	"path/filepath"
)

func getSystemDirectories(name string) ([]string, error) {
	userHome, err := getHomeDirectory()
	if err != nil {
		return []string{}, err
	}
	return []string{
		filepath.Join("/", "etc"),
		filepath.Join("/", "Library", "Application Support"),
		filepath.Join(userHome, "Library", "Application Support"),
	}, nil
}
