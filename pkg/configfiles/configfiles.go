package configfiles

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindConfigFiles provides a list of paths on the file system that are
// possible candidates for valid config files.
func FindConfigFiles(appname string, extensions []string) ([]string, error) {
	var configFiles []string

	directories, err := FindConfigDirectories(appname)
	if err != nil {
		return configFiles, err
	}

	for _, directory := range directories {
		for _, prefix := range []string{"", "."} {
			for _, suffix := range extensions {
				filename := fmt.Sprintf("%s%s.%s", prefix, appname, suffix)
				path, err := filepath.Abs(filepath.Join(directory, filename))
				if err != nil {
					continue
				}
				if _, err := os.Stat(path); os.IsNotExist(err) {
					continue
				}
				configFiles = append(configFiles, path)
			}
		}
	}

	return configFiles, nil
}
