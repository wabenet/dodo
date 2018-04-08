package command

import (
	"github.com/docker/docker/client"
	"github.com/oclaussen/dodo/pkg/config"
	"github.com/oclaussen/dodo/pkg/container"
	"github.com/oclaussen/dodo/pkg/image"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

// TODO: missing environment, user, volumes, volumes_from
// TODO: go through options of docker, docker-compose and sudo
type options struct {
	Filename    string
	Debug       bool
	Interactive bool
	NoCache     bool
	Pull        bool
	Build       bool
	Remove      bool
	NoRemove    bool
	Workdir     string
}

// TODO: no error message when bind mount fails

// TODO: tests

// NewCommand creates a new command instance
func NewCommand() *cobra.Command {
	var opts options
	cmd := &cobra.Command{
		Use:              "dodo [OPTIONS] NAME [CMD...]",
		Short:            "Run commands in a Docker context",
		TraverseChildren: true,
		Args:             cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommand(&opts, args[0], args[1:])
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Filename, "file", "f", "", "Specify a dodo configuration file")
	flags.BoolVarP(&opts.Debug, "debug", "", false, "Show additional debug output")
	flags.BoolVarP(&opts.Interactive, "interactive", "i", false, "Run an interactive session")
	flags.BoolVarP(&opts.NoCache, "no-cache", "", false, "Do not use cache when building the image")
	flags.BoolVarP(&opts.Pull, "pull", "", false, "Always attempt to pull a newer version of the image")
	flags.BoolVarP(&opts.Build, "build", "", false, "Always build an image, even if already exists")
	flags.BoolVarP(&opts.Remove, "rm", "", false, "Automatically remove the container when it exits")
	flags.BoolVarP(&opts.NoRemove, "no-rm", "", false, "Keep the container after it exits")
	flags.StringVarP(&opts.Workdir, "workdir", "w", "", "Working directory inside the container")
	flags.SetInterspersed(false)

	return cmd
}

func runCommand(options *options, name string, command []string) error {
	if options.Debug {
		// TODO: this does not seem to work?
		log.SetLevel(log.DebugLevel)
	}

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

	// TODO: generate a temp file in the container for the entrypoint
	// TODO feels inefficient to stupid all of config
	containerOptions := containerOptions(options, config)
	containerOptions.Client = dockerClient
	containerOptions.Image = imageID
	if len(command) > 0 {
		containerOptions.Command = command
	}
	return container.Run(ctx, containerOptions)
}

func imageOptions(options *options, config *config.BackdropConfig) image.Options {
	result := image.Options{
		Name:      config.Image,
		Build:     config.Build,
		ForcePull: config.Pull,
	}
	if options.Pull {
		result.ForcePull = options.Pull
	}
	if config.Build != nil && options.NoCache {
		result.Build.NoCache = true
	}
	if config.Build != nil && options.Build {
		result.Build.ForceRebuild = true
	}
	return result
}

func containerOptions(options *options, config *config.BackdropConfig) container.Options {
	result := container.Options{
		Name:        config.ContainerName,
		Interactive: config.Interactive,
		Remove:      true,
		Interpreter: config.Interpreter,
		Entrypoint:  "/tmp/dodo-entrypoint",
		Script:      config.Script,
		Command:     config.Command,
		Environment: config.Environment,
		Volumes:     config.Volumes,
		VolumesFrom: config.VolumesFrom,
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
	return result
}
