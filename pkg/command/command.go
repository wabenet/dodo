package command

import (
	"errors"
	"io/ioutil"

	"github.com/docker/cli/cli/config"
	"github.com/docker/docker/client"
	"github.com/oclaussen/dodo/pkg/container"
	"github.com/oclaussen/dodo/pkg/image"
	"github.com/spf13/cobra"
)

const description = `Run commands in a Docker context.

Dodo operates on a set of backdrops, that must be configured in configuration
files (in the current directory or one of the config directories). Backdrops
are similar to docker-composes services, but they define one-shot commands
instead of long-running services. More specifically, each backdrop defines a 
docker container in which a script should be executed. Dodo simply passes all 
CMD arguments to the first backdrop with NAME that is found. Additional FLAGS
can be used to overwrite the backdrop configuration.
`

// NewCommand creates a new command instance
func NewCommand() *cobra.Command {
	var opts options

	cmd := &cobra.Command{
		Use:                   "dodo [FLAGS] NAME [CMD...]",
		Short:                 "Run commands in a Docker context",
		Long:                  description,
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		TraverseChildren:      true,
		Args:                  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.list {
				return ListConfigurations()
			}
			if len(args) < 1 {
				return errors.New("please specify a backdrop name")
			}
			return runCommand(&opts, args[0], args[1:])
		},
	}

	opts.createFlags(cmd)
	return cmd
}

func runCommand(opts *options, name string, command []string) error {
	conf, err := LoadConfiguration(name, opts.file)
	if err != nil {
		return err
	}

	optsConfig, err := opts.createConfig(command)
	if err != nil {
		return err
	}

	conf.Merge(optsConfig)

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithVersion("1.39"))
	if err != nil {
		return err
	}
	authConfigs := config.LoadDefaultConfigFile(ioutil.Discard).GetAuthConfigs()

	image, err := image.NewImage(dockerClient, authConfigs, conf.Image)
	if err != nil {
		return err
	}
	imageID, err := image.Build()
	if err != nil {
		return err
	}

	container, err := container.NewContainer(dockerClient, conf)
	if err != nil {
		return err
	}

	return container.Run(imageID)
}
