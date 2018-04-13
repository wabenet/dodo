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

// TODO: go through options of docker, docker-compose and sudo
type options struct {
	Filename    string
	Quiet       bool
	Debug       bool
	Interactive bool
	NoCache     bool
	Pull        bool
	Build       bool
	Remove      bool
	NoRemove    bool
	Workdir     string
	User        string
	Volumes     []string
	VolumesFrom []string
	Environment []string
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
			initLogging(&opts)
			return runCommand(&opts, args[0], args[1:])
		},
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)

	flags.StringVarP(
		&opts.Filename, "file", "f", "",
		"specify a dodo configuration file")
	flags.BoolVarP(
		&opts.Quiet, "quiet", "q", false,
		"suppress informational output",
	)
	flags.BoolVarP(
		&opts.Debug, "debug", "", false,
		"show additional debug output")
	flags.BoolVarP(
		&opts.Interactive, "interactive", "i", false,
		"run an interactive session")
	flags.BoolVarP(
		&opts.NoCache, "no-cache", "", false,
		"do not use cache when building the image")
	flags.BoolVarP(
		&opts.Pull, "pull", "", false,
		"always attempt to pull a newer version of the image")
	flags.BoolVarP(
		&opts.Build, "build", "", false,
		"always build an image, even if already exists")
	flags.BoolVarP(
		&opts.Remove, "rm", "", false,
		"automatically remove the container when it exits")
	flags.BoolVarP(
		&opts.NoRemove, "no-rm", "", false,
		"keep the container after it exits")
	flags.StringVarP(
		&opts.Workdir, "workdir", "w", "",
		"working directory inside the container")
	flags.StringVarP(
		&opts.User, "user", "u", "",
		"Username or UID (format: <name|uid>[:<group|gid>])")
	flags.StringArrayVarP(
		&opts.Volumes, "volume", "v", []string{},
		"Bind mount a volume")
	flags.StringArrayVarP(
		&opts.VolumesFrom, "volumes-from", "", []string{},
		"Mount volumes from the specified container(s)")
	flags.StringArrayVarP(
		&opts.Environment, "env", "e", []string{},
		"Set environment variables")

	return cmd
}

func initLogging(options *options) {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
	})
	if options.Quiet {
		log.SetLevel(log.WarnLevel)
	} else if options.Debug {
		log.SetLevel(log.DebugLevel)
	}
}

func runCommand(options *options, name string, command []string) error {
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
	containerOptions := containerOptions(options, config)
	containerOptions.Client = dockerClient
	containerOptions.Image = imageID
	if len(command) > 0 {
		containerOptions.Command = command
	}
	return container.Run(ctx, containerOptions)
}

func imageOptions(
	options *options,
	config *config.BackdropConfig,
) image.Options {
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

func containerOptions(
	options *options,
	config *config.BackdropConfig,
) container.Options {
	result := container.Options{
		Name:        config.ContainerName,
		Interactive: config.Interactive,
		Remove:      true,
		Interpreter: config.Interpreter,
		Entrypoint:  "/tmp/dodo-entrypoint",
		Script:      config.Script,
		Command:     config.Command,
		Environment: append(config.Environment, options.Environment...),
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
