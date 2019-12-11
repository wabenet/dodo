package command

import (
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/spf13/cobra"
)

type options struct {
	interactive  bool
	remove       bool
	noRemove     bool
	build        bool
	noCache      bool
	pull         bool
	stage        string
	forwardStage bool
	user         string
	workdir      string
	volumes      []string
	volumesFrom  []string
	environment  []string
	publish      []string
}

func (opts *options) createFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.SetInterspersed(false)

	flags.BoolVarP(
		&opts.interactive, "interactive", "i", false,
		"run an interactive session")
	flags.BoolVarP(
		&opts.remove, "rm", "", false,
		"automatically remove the container when it exits")
	flags.BoolVarP(
		&opts.noRemove, "no-rm", "", false,
		"keep the container after it exits")
	flags.BoolVarP(
		&opts.build, "build", "", false,
		"always build an image, even if already exists")
	flags.BoolVarP(
		&opts.noCache, "no-cache", "", false,
		"do not use cache when building the image")
	flags.BoolVarP(
		&opts.pull, "pull", "", false,
		"always attempt to pull a newer version of the image")
	flags.StringVarP(
		&opts.stage, "stage", "s", "",
		"stage to user for docker daemon")
	flags.BoolVarP(
		&opts.forwardStage, "forward-stage", "", false,
		"forward stage information into container, so dodo can be used inside the container")
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
		&opts.volumesFrom, "volumes-from", "", []string{},
		"mount volumes from the specified container(s)")
	flags.StringArrayVarP(
		&opts.environment, "env", "e", []string{},
		"set environment variables")
	flags.StringArrayVarP(
		&opts.publish, "publish", "p", []string{},
		"publish a container's port(s) to the host")
}

func (opts *options) createConfig(command []string) (*types.Backdrop, error) {
	config := &types.Backdrop{
		Image: &types.Image{
			ForceRebuild: opts.build,
			NoCache:      opts.noCache,
			ForcePull:    opts.pull,
		},
		Interactive:  opts.interactive,
		Stage:        opts.stage,
		ForwardStage: opts.forwardStage,
		User:         opts.user,
		WorkingDir:   opts.workdir,
		VolumesFrom:  opts.volumesFrom,
		Command:      command,
	}

	if opts.noRemove {
		remove := false
		config.Remove = &remove
	}
	if opts.remove {
		remove := true
		config.Remove = &remove
	}

	decoder := types.NewDecoder("cli")

	for _, volume := range opts.volumes {
		decoded, err := decoder.DecodeVolume("cli", volume)
		if err != nil {
			return nil, err
		}
		config.Volumes = append(config.Volumes, decoded)
	}

	for _, env := range opts.environment {
		decoded, err := decoder.DecodeKeyValue("cli", env)
		if err != nil {
			return nil, err
		}
		config.Environment = append(config.Environment, decoded)
	}

	for _, port := range opts.publish {
		decoded, err := decoder.DecodePort("cli", port)
		if err != nil {
			return nil, err
		}
		config.Ports = append(config.Ports, decoded)
	}

	return config, nil
}
