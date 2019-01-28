package command

import (
	"errors"
	"io/ioutil"

	cliconfig "github.com/docker/cli/cli/config"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stringid"
	"github.com/oclaussen/dodo/pkg/config"
	"github.com/oclaussen/dodo/pkg/container"
	"github.com/oclaussen/dodo/pkg/image"
	"github.com/oclaussen/dodo/pkg/logging"
	"github.com/oclaussen/dodo/pkg/options"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
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
	var loggingOpts logging.Options
	var dodoOpts options.Options

	cmd := &cobra.Command{
		Use:                   "dodo [FLAGS] NAME [CMD...]",
		Short:                 "Run commands in a Docker context",
		Long:                  description,
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		TraverseChildren:      true,
		Args:                  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			logging.InitLogging(&loggingOpts)
			if dodoOpts.List {
				config.ListConfigurations()
				return nil
			}
			if len(args) < 1 {
				return errors.New("Please specify a backdrop name")
			}
			return runCommand(&dodoOpts, args[0], args[1:])
		},
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)
	logging.InitFlags(flags, &loggingOpts)
	options.InitFlags(flags, &dodoOpts)
	return cmd
}

func runCommand(options *options.Options, name string, command []string) error {
	config := config.LoadConfiguration(name, options.Filename)
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	// TODO: log errors
	authConfigs := cliconfig.LoadDefaultConfigFile(ioutil.Discard).GetAuthConfigs()

	ctx := context.Background()

	imageOptions := imageOptions(options, config)
	imageOptions.Client = dockerClient
	imageOptions.AuthConfigs = authConfigs
	imageID, err := image.Get(imageOptions)
	if err != nil {
		return err
	}

	containerOptions := containerOptions(options, config, command)
	containerOptions.Client = dockerClient
	containerOptions.Image = imageID
	return container.Run(ctx, containerOptions)
}

func imageOptions(
	options *options.Options,
	config *config.BackdropConfig,
) image.Options {
	result := image.Options{
		Name:       config.Build.Name,
		ForcePull:  config.Build.ForcePull,
		NoCache:    config.Build.NoCache,
		Context:    config.Build.Context,
		Dockerfile: config.Build.Dockerfile,
		Steps:      config.Build.Steps,
		Args:       config.Build.Args.Strings(),
		ForceBuild: config.Build.ForceRebuild,
	}
	if options.Pull {
		result.ForcePull = options.Pull
	}
	if options.Build {
		result.ForceBuild = true
	}
	if options.NoCache {
		result.NoCache = true
	}
	return result
}

func containerOptions(
	options *options.Options,
	config *config.BackdropConfig,
	command []string,
) container.Options {
	result := container.Options{
		Name:        config.ContainerName,
		Remove:      true,
		Entrypoint:  []string{"/bin/sh"},
		Script:      config.Script,
		ScriptPath:  "/tmp/dodo-dockerfile-" + stringid.GenerateRandomID()[:20],
		Command:     config.Command,
		Environment: append(config.Environment.Strings(), options.Environment...),
		Volumes:     append(config.Volumes.Strings(), options.Volumes...),
		VolumesFrom: append(config.VolumesFrom, options.VolumesFrom...),
		User:        config.User,
		WorkingDir:  config.WorkingDir,
	}
	if options.Workdir != "" {
		result.WorkingDir = options.Workdir
	}
	if config.Remove != nil {
		result.Remove = *config.Remove
	}
	if options.Remove {
		result.Remove = true
	}
	if options.NoRemove {
		result.Remove = false
	}
	if options.User != "" {
		result.User = options.User
	}

	if config.Interpreter != nil {
		result.Entrypoint = config.Interpreter
	}
	if config.Interactive || options.Interactive {
		result.Command = nil
	} else {
		if len(config.Script) > 0 {
			result.Entrypoint = append(result.Entrypoint, result.ScriptPath)
		}
		if len(command) > 0 {
			result.Command = command
		}
	}

	return result
}
