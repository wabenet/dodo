package run

import (
	"fmt"

	"github.com/spf13/cobra"
	api "github.com/wabenet/dodo-core/api/v1alpha4"
	"github.com/wabenet/dodo-core/pkg/config"
	"github.com/wabenet/dodo-core/pkg/plugin"
	"github.com/wabenet/dodo-core/pkg/plugin/command"
	"github.com/wabenet/dodo/pkg/core"
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

type options struct {
	interactive bool
	user        string
	workdir     string
	volumes     []string
	environment []string
	publish     []string
	runtime     string
}

func New(m plugin.Manager) *Command {
	var opts options

	cmd := &cobra.Command{
		Use:                   Name,
		Short:                 "Run commands in Docker context",
		Long:                  description,
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Args:                  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			backdrop, err := opts.createConfig(args[0], args[1:])
			if err != nil {
				return fmt.Errorf("error running backdrop: %w", err)
			}

			exitCode, err := core.RunByName(m, backdrop)
			command.SetExitCode(cmd, exitCode)

			if err != nil {
				return fmt.Errorf("error running backdrop: %w", err)
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)

	flags.BoolVarP(
		&opts.interactive, "interactive", "i", false,
		"run an interactive session")
	flags.StringVarP(
		&opts.user, "user", "u", "",
		"username or UID (format: <name|uid>[:<group|gid>])")
	flags.StringVarP(
		&opts.workdir, "workdir", "w", "",
		"working directory inside the container")
	flags.StringArrayVarP(
		&opts.volumes, "volume", "v", []string{},
		"bind mount a volume")
	flags.StringArrayVarP(
		&opts.environment, "env", "e", []string{},
		"set environment variables")
	flags.StringArrayVarP(
		&opts.publish, "publish", "p", []string{},
		"publish a container's port(s) to the host")
	flags.StringVarP(
		&opts.runtime, "runtime", "r", "",
		"select runtime plugin")

	return &Command{cmd: cmd}
}

func (opts *options) createConfig(name string, command []string) (*api.Backdrop, error) {
	c := &api.Backdrop{
		Name:    name,
		Runtime: opts.runtime,
		Entrypoint: &api.Entrypoint{
			Interactive: opts.interactive,
			Arguments:   command,
		},
		User:       opts.user,
		WorkingDir: opts.workdir,
	}

	for _, spec := range opts.volumes {
		vol, err := config.ParseVolumeMount(spec)
		if err != nil {
			return nil, fmt.Errorf("could not parse volume config: %w", err)
		}

		c.Volumes = append(c.Volumes, vol)
	}

	for _, spec := range opts.environment {
		env, err := config.ParseEnvironmentVariable(spec)
		if err != nil {
			return nil, fmt.Errorf("could not parse environment config: %w", err)
		}

		c.Environment = append(c.Environment, env)
	}

	for _, spec := range opts.publish {
		port, err := config.ParsePortBinding(spec)
		if err != nil {
			return nil, fmt.Errorf("could not parse port config: %w", err)
		}

		c.Ports = append(c.Ports, port)
	}

	return c, nil
}
