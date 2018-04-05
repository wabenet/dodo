package options

import (
	"github.com/oclaussen/dodo/config"
)

// TODO: add all all the options
// TODO: add some --no-rm option?
type Options struct {
	Filename    string
	Remove      bool
	NoCache     bool
	Pull        bool
	Build       bool
}

func (options *Options) UpdateConfiguration(config *config.ContextConfig) {
	if options.Remove {
		remove := true
		config.Remove = &remove
	}
	if config.Build != nil && options.NoCache {
		config.Build.NoCache = true
	}
	if options.Pull {
		config.Pull = true
	}
	if config.Build != nil && options.Build {
		config.Build.ForceRebuild = true
	}
}
