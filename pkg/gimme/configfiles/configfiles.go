package configfiles

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type ConfigFile struct {
	Path    string
	Content []byte
}

func GimmeConfigFiles(opts *Options) (*ConfigFile, error) {
	if err := opts.normalize(); err != nil {
		return nil, err
	}

	for _, pattern := range opts.FileGlobs {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			// TODO print warning
			continue
		}
		for _, file := range matches {
			configFile, ok := tryConfigFile(file)
			if ok && opts.Filter(configFile) {
				return configFile, nil
			}
		}
	}

	if opts.UseFileGlobsOnly {
		return nil, errors.New("no matching configuration file found")
	}

	directories, err := GimmeConfigDirectories(opts)
	if err != nil {
		return nil, err
	}

	for _, directory := range directories {
		for _, prefix := range []string{"", "."} {
			for _, suffix := range opts.Extensions {
				var filename string
				if len(suffix) > 0 {
					filename = fmt.Sprintf("%s%s.%s", prefix, opts.Name, suffix)
				} else {
					filename = fmt.Sprintf("%s%s", prefix, opts.Name)
				}
				configFile, ok := tryConfigFile(filepath.Join(directory, filename))
				if ok && opts.Filter(configFile) {
					return configFile, nil
				}
			}
		}
	}

	return nil, errors.New("no matching configuration file found")
}

func tryConfigFile(path string) (*ConfigFile, bool) {
	var err error
	result := &ConfigFile{}

	if result.Path, err = filepath.Abs(path); err != nil {
		return result, false
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return result, false
	}

	if result.Content, err = ioutil.ReadFile(path); err != nil {
		return result, false
	}

	return result, true
}
