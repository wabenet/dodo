package options

import (
	"github.com/spf13/pflag"
)

// TODO: go through options of docker, docker-compose and sudo

// Options represents the set of command-line options of the command
type Options struct {
	Filename    string
	List        bool
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

// InitFlags adds flags for all possible command-line options to a flag set
func InitFlags(flags *pflag.FlagSet, opts *Options) {
	flags.StringVarP(
		&opts.Filename, "file", "f", "",
		"specify a dodo configuration file")
	flags.BoolVarP(
		&opts.List, "list", "", false,
		"list all available backdrop configurations")

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
}
