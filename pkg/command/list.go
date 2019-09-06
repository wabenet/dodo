package command

import (
	"fmt"

	"github.com/oclaussen/dodo/pkg/types"
	"github.com/oclaussen/go-gimme/configfiles"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available all backdrop configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
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
				return false
			}
			decoder := types.NewDecoder(configFile.Path, "")
			if config, err := decoder.DecodeNames(configFile.Path, "", mapType); err == nil {
				names.Merge(&config)
			}
			return false
		},
	})
	for _, item := range names.Strings() {
		fmt.Printf("%s\n", item)
	}
	return nil
}
