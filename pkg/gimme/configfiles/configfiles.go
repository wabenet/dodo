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

	patterns := filenameGlobs(opts.Name, opts.Extensions)
	for _, directory := range directories {
		for _, pattern := range patterns {
			matches, err := filepath.Glob(filepath.Join(directory, pattern))
			if err != nil {
				// TODO print warning
				continue
			}
			for _, filename := range matches {
				configFile, ok := tryConfigFile(filename)
				if ok && opts.Filter(configFile) {
					return configFile, nil
				}
			}
		}
	}

	return nil, errors.New("no matching configuration file found")
}

func filenameGlobs(name string, extensions []string) []string {
	if len(extensions) == 0 {
		return []string{
			name,
			fmt.Sprintf(".%s", name),
			filepath.Join(name, "config"),
			filepath.Join(fmt.Sprintf(".%s", name), "config"),
			filepath.Join(fmt.Sprintf("%s.d", name), "*"),
			filepath.Join(fmt.Sprintf(".%s.d", name), "*"),
		}
	}

	var candidates []string
	for _, ext := range extensions {
		candidates = append(candidates, []string{
			fmt.Sprintf("%s.%s", name, ext),
			fmt.Sprintf(".%s.%s", name, ext),
			filepath.Join(name, fmt.Sprintf("config.%s", ext)),
			filepath.Join(fmt.Sprintf(".%s", name), fmt.Sprintf("config.%s", ext)),
			filepath.Join(fmt.Sprintf("%s.%s.d", name, ext), "*"),
			filepath.Join(fmt.Sprintf(".%s.%s.d", name, ext), "*"),
		}...)
	}
	return candidates
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
