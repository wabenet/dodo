package config

import (
	"fmt"

	"github.com/oclaussen/dodo/pkg/types"
	"github.com/oclaussen/go-gimme/configfiles"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func LoadConfiguration(backdrop string, filename string) (*types.Backdrop, error) {
	var opts *configfiles.Options
	if len(filename) > 0 {
		opts = &configfiles.Options{
			FileGlobs:        []string{filename},
			UseFileGlobsOnly: true,
		}
	} else {
		opts = &configfiles.Options{
			Name:                      "dodo",
			Extensions:                []string{"yaml", "yml", "json"},
			IncludeWorkingDirectories: true,
			Filter: func(configFile *configfiles.ConfigFile) bool {
				return containsBackdrop(configFile, backdrop)
			},
		}
	}

	configFile, err := configfiles.GimmeConfigFiles(opts)
	if err != nil {
		return nil, err
	}

	var mapType map[interface{}]interface{}
	if err := yaml.Unmarshal(configFile.Content, &mapType); err != nil {
		return nil, err
	}

	decoder := types.NewDecoder(configFile.Path, backdrop)
	config, err := decoder.DecodeGroup(configFile.Path, mapType)
	if err != nil {
		return nil, err
	}

	if result, ok := config.Backdrops[backdrop]; ok {
		return &result, nil
	}

	return nil, fmt.Errorf("could not find backdrop %s in file %s", backdrop, configFile.Path)
}

func LoadNames() *types.Names {
	names := types.Names{}
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
			decoder := types.NewDecoder(configFile.Path, "")
			config, err := decoder.DecodeNames(configFile.Path, "", mapType)
			if err != nil {
				log.WithFields(log.Fields{"file": configFile.Path, "reason": err}).Warn("invalid config file")
				return false
			}
			names.Merge(&config)
			return false
		},
	})
	return &names
}

func ValidateConfigs(files []string) error {
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

			decoder := types.NewDecoder(configFile.Path, "")
			if _, err := decoder.DecodeNames(configFile.Path, "", mapType); err != nil {
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

func containsBackdrop(configFile *configfiles.ConfigFile, backdrop string) bool {
	var mapType map[interface{}]interface{}
	if err := yaml.Unmarshal(configFile.Content, &mapType); err != nil {
		log.WithFields(log.Fields{"file": configFile.Path}).Warn("invalid YAML syntax in file")
		return false
	}

	decoder := types.NewDecoder(configFile.Path, backdrop)
	config, err := decoder.DecodeNames(configFile.Path, "", mapType)
	if err != nil {
		log.WithFields(log.Fields{"file": configFile.Path, "reason": err}).Warn("invalid config file")
		return false
	}

	_, ok := config.Backdrops[backdrop]
	return ok
}
