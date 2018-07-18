package configfiles

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type finder struct {
	AppName        string
	FileExtensions []string
	Process        func(string) bool
}

func NewFinder(
	appname string, extensions []string, process func(string) bool,
) *finder {
	return &finder{
		AppName:        appname,
		FileExtensions: extensions,
		Process:        process,
	}
}

func (finder *finder) FindConfiguration() error {
	directories, err := FindConfigDirectories(finder.AppName)
	if err != nil {
		return err
	}

	for _, directory := range directories {
		found := finder.handleDirectory(directory)
		if found {
			return nil
		}
	}

	return fmt.Errorf("Could not find any configuration for %s", finder.AppName)
}

func (finder *finder) handleDirectory(directory string) bool {
	for _, prefix := range []string{"", "."} {
		for _, suffix := range finder.FileExtensions {
			filename := fmt.Sprintf("%s%s.%s", prefix, finder.AppName, suffix)
			path, err := filepath.Abs(filepath.Join(directory, filename))
			if err != nil {
				log.Error(err)
			}
			if _, err := os.Stat(path); os.IsNotExist(err) {
				continue
			}
			if found := finder.Process(path); found {
				return found
			}
		}
	}
	return false
}
