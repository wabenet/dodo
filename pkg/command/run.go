package command

import (
	"github.com/docker/docker/client"
	"github.com/oclaussen/dodo/pkg/container"
	"github.com/oclaussen/dodo/pkg/image"
	"github.com/spf13/cobra"
)

func NewRunCommand() *cobra.Command {
	var opts options
	cmd := &cobra.Command{
		Use:                   "run [flags] [name] [cmd...]",
		Short:                 "Same as running 'dodo [name]', can be used when a backdrop name collides with a top-level command",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Args:                  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

	image, err := image.NewImage(dockerClient, LoadAuthConfig(), conf.Image)
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
