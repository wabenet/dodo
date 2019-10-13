package config

import (
	"fmt"

	"github.com/oclaussen/dodo/pkg/types"
	"github.com/oclaussen/go-gimme/configfiles"
	"github.com/sahilm/fuzzy"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func LoadBackdrop(backdrop string) (*types.Backdrop, error) {
	config := loadConfig()
	if result, ok := config.Backdrops[backdrop]; ok {
		return &result, nil
	}

	matches := fuzzy.Find(backdrop, config.Names())
	if len(matches) == 0 {
		return nil, fmt.Errorf("could not find any configuration for backdrop '%s'", backdrop)
	}
	return nil, fmt.Errorf("backdrop '%s' not found, did you mean '%s'?", backdrop, matches[0].Str)
}

func LoadImage(image string) (*types.Image, error) {
	config := loadConfig()
	for _, backdrop := range config.Backdrops {
		if backdrop.Image != nil && backdrop.Image.Name == image {
			return backdrop.Image, nil
		}
	}
	return nil, fmt.Errorf("could not find any backdrop configuration that would produce image '%s'", image)
}

func LoadStage(name string) (*types.Stage, error) {
	config := loadConfig()
	if result, ok := config.Stages[name]; ok {
		return &result, nil
	}
	return nil, fmt.Errorf("could not find any configuration for stage '%s'", name)
}

func ListBackdrops() []string {
	return loadConfig().Strings()
}

func ValidateConfigs(files []string) error {
	// TODO: too much duplication between load and validate
	errors := 0
	configfiles.GimmeConfigFiles(&configfiles.Options{
		FileGlobs:        files,
		UseFileGlobsOnly: true,
		Filter: func(configFile *configfiles.ConfigFile) bool {
			var mapType map[interface{}]interface{}
			if err := yaml.Unmarshal(configFile.Content, &mapType); err != nil {
				log.WithFields(log.Fields{"file": configFile.Path}).Error("invalid YAML syntax in file")
				errors = errors + 1
				return false
			}

			decoder := types.NewDecoder(configFile.Path)
			if _, err := decoder.DecodeGroup(configFile.Path, mapType); err != nil {
				log.WithFields(log.Fields{"file": configFile.Path, "reason": err}).Error("invalid config file")
				errors = errors + 1
				return false
			}

			log.WithFields(log.Fields{"file": configFile.Path}).Info("config file ok")
			return false
		},
	})
	if errors > 0 {
		return fmt.Errorf("%d errors encountered", errors)
	}
	return nil
}

func loadConfig() *types.Group {
	var result types.Group
	configfiles.GimmeConfigFiles(&configfiles.Options{
		Name:                      "dodo",
		Extensions:                []string{"yaml", "yml", "json"},
		IncludeWorkingDirectories: true,
		Filter: func(configFile *configfiles.ConfigFile) bool {
			var mapType map[interface{}]interface{}
			if err := yaml.Unmarshal(configFile.Content, &mapType); err != nil {
				log.WithFields(log.Fields{"file": configFile.Path}).Warn("invalid YAML syntax in file")
				return false
			}

			decoder := types.NewDecoder(configFile.Path)
			config, err := decoder.DecodeGroup(configFile.Path, mapType)
			if err != nil {
				log.WithFields(log.Fields{"file": configFile.Path, "reason": err}).Warn("invalid config file")
				return false
			}

			result.Merge(&config)
			return false
		},
	})
	return &result
}
