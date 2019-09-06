package command

import (
	"fmt"

	"github.com/oclaussen/dodo/pkg/types"
	"github.com/oclaussen/go-gimme/configfiles"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available all backdrop configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			configureLogging()
			return listConfigurations()
		},
	}
}

func listConfigurations() error {
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
	for _, item := range names.Strings() {
		fmt.Printf("%s\n", item)
	}
	return nil
}
