package command

import (
	"fmt"

	"github.com/oclaussen/dodo/pkg/stage"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/oclaussen/go-gimme/configfiles"
	"github.com/oclaussen/go-gimme/ssh"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func NewStageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "stage",
		Short:            "Manage stages",
		TraverseChildren: true,
		SilenceUsage:     true,
	}

	cmd.AddCommand(NewUpCommand())
	cmd.AddCommand(NewDownCommand())
	cmd.AddCommand(NewSSHCommand())
	return cmd
}

func NewUpCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Create or start a stage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := loadStageConfig(args[0])
			if err != nil {
				return err
			}
			s, cleanup, err := stage.Load(args[0], conf)
			defer cleanup()
			if err != nil {
				return err
			}

			exist, err := s.Exist()
			if err != nil {
				return err
			}

			if !exist {
				return s.Create()
			}
			return s.Start()
		},
	}
}

type downOptions struct {
	remove bool
	force  bool
}

func NewDownCommand() *cobra.Command {
	var opts downOptions
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Remove or pause a stage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: do we actually need the config?
			conf, err := loadStageConfig(args[0])
			if err != nil {
				return err
			}
			s, cleanup, err := stage.Load(args[0], conf)
			defer cleanup()
			if err != nil {
				return err
			}

			if opts.remove {
				return s.Remove(opts.force)
			}
			return s.Stop()
		},
	}
	flags := cmd.Flags()
	flags.BoolVarP(&opts.remove, "rm", "", false, "remove the stage instead of pausing")
	flags.BoolVarP(&opts.force, "force", "f", false, "when used with '--rm', don't stop on errors")
	return cmd
}

func NewSSHCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ssh",
		Short: "login to the stage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: do we actually need the config?
			conf, err := loadStageConfig(args[0])
			if err != nil {
				return err
			}
			s, cleanup, err := stage.Load(args[0], conf)
			defer cleanup()
			if err != nil {
				return err
			}

			available, err := s.Available()
			if err != nil {
				return err
			}

			if !available {
				return errors.New("stage is not up")
			}

			opts, err := s.GetSSHOptions()
			if err != nil {
				return err
			}

			return ssh.GimmeShell(&ssh.Options{
				Host:              opts.Hostname,
				Port:              opts.Port,
				User:              opts.Username,
				IdentityFileGlobs: []string{opts.PrivateKeyFile},
				NonInteractive:    true,
			})
		},
	}
}

func loadStageConfig(name string) (*types.Stage, error) {
	configFile, err := configfiles.GimmeConfigFiles(&configfiles.Options{
		Name:                      "dodo",
		Extensions:                []string{"yaml", "yml", "json"},
		IncludeWorkingDirectories: true,
		Filter: func(configFile *configfiles.ConfigFile) bool {
			return containsStage(configFile, name)
		},
	})
	if err != nil {
		return nil, err
	}

	var mapType map[interface{}]interface{}
	if err := yaml.Unmarshal(configFile.Content, &mapType); err != nil {
		return nil, err
	}

	config, err := types.DecodeGroup(configFile.Path, mapType)
	if err != nil {
		return nil, err
	}

	if result, ok := config.Stages[name]; ok {
		return &result, nil
	}

	return nil, fmt.Errorf("could not find stage %s in file %s", name, configFile.Path)
}

func containsStage(configFile *configfiles.ConfigFile, name string) bool {
	var mapType map[interface{}]interface{}
	if err := yaml.Unmarshal(configFile.Content, &mapType); err != nil {
		return false
	}

	config, err := types.DecodeNames(configFile.Path, "", mapType)
	if err != nil {
		return false
	}

	_, ok := config.Stages[name]
	return ok
}
