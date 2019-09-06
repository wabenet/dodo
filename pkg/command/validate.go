package command

import (
	"fmt"

	"github.com/oclaussen/dodo/pkg/types"
	"github.com/oclaussen/go-gimme/configfiles"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func NewValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:                   "validate",
		Short:                 "Validate configuration files for syntax errors",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Args:                  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configureLogging()
			return validateConfigs(args)
		},
	}
}

func validateConfigs(files []string) error {
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
