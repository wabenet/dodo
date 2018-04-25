package command

import (
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stringid"
	"github.com/oclaussen/dodo/pkg/config"
	"github.com/oclaussen/dodo/pkg/container"
	"github.com/oclaussen/dodo/pkg/image"
	"github.com/oclaussen/dodo/pkg/options"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

// TODO: no error message when bind mount fails

// TODO: tests

// NewCommand creates a new command instance
func NewCommand() *cobra.Command {
	var opts options.Options
	cmd := &cobra.Command{
		Use:              "dodo [OPTIONS] NAME [CMD...]",
		Short:            "Run commands in a Docker context",
		SilenceUsage:     true,
		TraverseChildren: true,
		Args:             cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			initLogging(&opts)
			return runCommand(&opts, args[0], args[1:])
		},
	}
	options.ConfigureFlags(cmd, &opts)
	return cmd
}

func initLogging(options *options.Options) {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
	})
	if options.Quiet {
		log.SetLevel(log.WarnLevel)
	} else if options.Debug {
		log.SetLevel(log.DebugLevel)
	}
}

func runCommand(options *options.Options, name string, command []string) error {
	config, err := config.LoadConfiguration(name, options.Filename)
	if err != nil {
		return err
	}

	// TODO: read docker configuration
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	ctx := context.Background()

	imageOptions := imageOptions(options, config)
	imageOptions.Client = dockerClient
	imageID, err := image.Get(ctx, imageOptions)
	if err != nil {
		return err
	}

	containerOptions := containerOptions(options, config)
	containerOptions.Client = dockerClient
	containerOptions.Image = imageID
	if len(command) > 0 {
		containerOptions.Command = command
	}
	return container.Run(ctx, containerOptions)
}

func imageOptions(
	options *options.Options,
	config *config.BackdropConfig,
) image.Options {
	result := image.Options{
		Name:      config.Image,
		ForcePull: config.Pull,
	}
	if options.Pull {
		result.ForcePull = options.Pull
	}
	if config.Build != nil {
		result.DoBuild = true
		result.Context = config.Build.Context
		result.Dockerfile = config.Build.Dockerfile
		result.Steps = config.Build.Steps
		result.Args = config.Build.Args.Strings()
		result.ForceBuild = config.Build.ForceRebuild
		if options.Build {
			result.ForceBuild = true
		}
		result.NoCache = config.Build.NoCache
		if options.NoCache {
			result.NoCache = true
		}
	}
	return result
}

func containerOptions(
	options *options.Options,
	config *config.BackdropConfig,
) container.Options {
	entrypoint := "/tmp/dodo-dockerfile-" + stringid.GenerateRandomID()[:20]
	result := container.Options{
		Name:        config.ContainerName,
		Interactive: config.Interactive,
		Remove:      true,
		Interpreter: config.Interpreter,
		Entrypoint:  entrypoint,
		Script:      config.Script,
		Command:     config.Command,
		Environment: append(config.Environment.Strings(), options.Environment...),
		Volumes:     append(config.Volumes, options.Volumes...),
		VolumesFrom: append(config.VolumesFrom, options.VolumesFrom...),
		User:        config.User,
		WorkingDir:  config.WorkingDir,
	}
	if options.Workdir != "" {
		result.WorkingDir = options.Workdir
	}
	if options.Interactive {
		result.Interactive = true
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
	return result
}
