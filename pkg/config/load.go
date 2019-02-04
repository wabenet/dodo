package config

import (
	"fmt"
	"os"
	"os/user"

	"github.com/oclaussen/dodo/pkg/configfiles"
	"github.com/oclaussen/dodo/pkg/image"
	"github.com/oclaussen/dodo/pkg/types"
	log "github.com/sirupsen/logrus"
)

// LoadConfiguration tries to find a backdrop configuration by name in any of
// the supported locations. If given, will only look in the supplied config
// file.
func LoadConfiguration(
	backdrop string, configfile string,
) *BackdropConfig {
	var result BackdropConfig
	processFile := func(filename string) bool {
		ok := false
		config, err := ParseConfigurationFile(filename)
		if err != nil {
			log.WithFields(log.Fields{
				"file":   filename,
				"reason": err,
			}).Error("Could not parse config file")
			return false
		}
		result, ok = config.Backdrops[backdrop]
		return ok
	}

	if configfile != "" {
		if found := processFile(configfile); found {
			return &result
		}
		log.WithFields(log.Fields{
			"name": backdrop,
			"file": configfile,
		}).Error("No valid config in file")
	}

	err := configfiles.NewFinder(
		"dodo",
		[]string{"yaml", "yml", "json"},
		processFile,
	).FindConfiguration()

	if err == nil {
		return &result
	}

	log.WithFields(log.Fields{
		"reason": err,
	}).Info("Fallback to default configuration")
	return FallbackConfig(backdrop)
}

// FallbackConfig guesses a general-purpose backdrop configuration based
// on the name, that can be used in case no better configuration was found.
func FallbackConfig(backdrop string) *BackdropConfig {
	result := &BackdropConfig{
		Image:  &image.ImageConfig{Steps: []string{fmt.Sprintf("FROM %s", backdrop)}},
		Script: fmt.Sprintf("%s $@", backdrop),
	}

	if workingDir, err := os.Getwd(); err != nil {
		result.WorkingDir = workingDir
		result.Volumes = []types.Volume{types.Volume{
			Source: workingDir,
			Target: workingDir,
		}}
	} else {
		log.Error(err)
	}

	if user, err := user.Current(); err != nil {
		result.User = fmt.Sprintf("%s:%s", user.Uid, user.Gid)
	} else {
		log.Error(err)
	}

	return result
}

// ListConfigurations prints out all available backdrop names and the file
// it was found in.
func ListConfigurations() {
	result := map[string]string{}
	configfiles.NewFinder(
		"dodo",
		[]string{"yaml", "yml", "json"},
		func(filename string) bool {
			config, err := ParseConfigurationFile(filename)
			if err != nil {
				return false
			}

			for name := range config.Backdrops {
				if result[name] == "" {
					log.WithFields(log.Fields{
						"file": filename,
					}).Info(name)
				}
				result[name] = filename
			}

			return false
		},
	).FindConfiguration()
}
