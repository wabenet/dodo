package options

import (
	"github.com/oclaussen/dodo/config"
	"github.com/spf13/cobra"
)

// TODO: add all all the options
// TODO: add some --no-rm option?
type Options struct {
	Filename    string
	Remove      bool
	NoCache     bool
	Pull        bool
}

func NewOptions() *Options {
	return &Options{}
}

func (options *Options) CreateFlags(command *cobra.Command) {
	opts := *options
	flags := command.Flags()

	flags.StringVarP(&opts.Filename, "file", "f", "", "Specify a dodo configuration file")
	flags.BoolVarP(&opts.Remove, "rm", "", false, "Automatically remove the container when it exits")
	flags.BoolVarP(&opts.NoCache, "no-cache", "", false, "Do not use cache when building the image")
	flags.BoolVarP(&opts.Pull, "pull", "", false, "Always attempt to pull a newer version of the image")

	flags.SetInterspersed(false)
}

func (options *Options) UpdateConfiguration(config *config.ContextConfig) {
	if options.Pull {
		config.Pull = true
	}
	if options.Remove {
		remove := true
		config.Remove = &remove
	}
	if config.Build != nil && options.NoCache {
		config.Build.NoCache = true
	}
}
