package command

import (
	"fmt"

	"github.com/oclaussen/dodo/pkg/config"
	"github.com/oclaussen/dodo/pkg/container"
	"github.com/oclaussen/dodo/pkg/stage"
	"github.com/oclaussen/dodo/pkg/stages/defaultchain"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// TODO: clean up command structure and options

const description = `Run commands in a Docker context.

Dodo operates on a set of backdrops, that must be configured in configuration
files (in the current directory or one of the config directories). Backdrops
are similar to docker-composes services, but they define one-shot commands
instead of long-running services. More specifically, each backdrop defines a 
docker container in which a script should be executed. Dodo simply passes all 
CMD arguments to the first backdrop with NAME that is found. Additional FLAGS
can be used to overwrite the backdrop configuration.
`

func NewCommand() *cobra.Command {
	var opts options
	cmd := &cobra.Command{
		Use:                   "dodo [flags] [name] [cmd...]",
		Short:                 "Run commands in a Docker context",
		Long:                  description,
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Args:                  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configureLogging()

			conf, err := config.LoadBackdrop(args[0])
			if err != nil {
				return err
			}

			optsConfig, err := opts.createConfig(args[1:])
			if err != nil {
				return err
			}

			conf.Merge(optsConfig)

			return withStage(conf.Stage, func(s stage.Stage) error {
				c, err := container.NewContainer(conf, s, config.LoadAuthConfig(), false)
				if err != nil {
					return err
				}

				return c.Run()
			})
		},
	}
	opts.createFlags(cmd)

	cmd.AddCommand(NewRunCommand())
	cmd.AddCommand(NewBuildCommand())
	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewValidateCommand())
	cmd.AddCommand(NewDaemonCommand())
	cmd.AddCommand(NewStageCommand())
	return cmd
}

func NewRunCommand() *cobra.Command {
	var opts options
	cmd := &cobra.Command{
		Use:                   "run [flags] [name] [cmd...]",
		Short:                 "Same as running 'dodo [name]', can be used when a backdrop name collides with a top-level command",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Args:                  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configureLogging()

			conf, err := config.LoadBackdrop(args[0])
			if err != nil {
				return err
			}

			optsConfig, err := opts.createConfig(args[1:])
			if err != nil {
				return err
			}

			conf.Merge(optsConfig)

			return withStage(conf.Stage, func(s stage.Stage) error {
				c, err := container.NewContainer(conf, s, config.LoadAuthConfig(), false)
				if err != nil {
					return err
				}

				return c.Run()
			})
		},
	}

	opts.createFlags(cmd)
	return cmd
}

func NewBuildCommand() *cobra.Command {
	return &cobra.Command{
		Use:                   "build",
		Short:                 "Build all required images for backdrop without running it",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Args:                  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configureLogging()
			conf, err := config.LoadBackdrop(args[0])
			if err != nil {
				return err
			}

			return withStage(conf.Stage, func(s stage.Stage) error {
				c, err := container.NewContainer(conf, s, config.LoadAuthConfig(), false)
				if err != nil {
					return err
				}

				return c.Build()
			})
		},
	}
}

func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:                   "list",
		Short:                 "List available all backdrop configurations",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		RunE: func(cmd *cobra.Command, args []string) error {
			configureLogging()
			for _, item := range config.ListBackdrops() {
				fmt.Printf("%s\n", item)
			}
			return nil
		},
	}
}

func NewValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:                   "validate",
		Short:                 "Validate configuration files for syntax errors",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Args:                  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configureLogging()
			return config.ValidateConfigs(args)
		},
	}
}

func configureLogging() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
	})
}

func withStage(name string, thing func(stage.Stage) error) error {
	var conf *types.Stage
	var err error

	if len(name) > 0 {
		conf, err = config.LoadStage(name)
		if err != nil {
			return err
		}
	}

	s := &defaultchain.Stage{}
	defer s.Cleanup()
	if err := s.Initialize(name, conf); err != nil {
		return errors.Wrap(err, "initialization failed")
	}

	return thing(s)
}
