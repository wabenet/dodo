package options

import (
	"github.com/oclaussen/dodo/config"
)

// TODO: add some --no-rm option?
// TODO: missing environment, user, volumes, volumes_from
// TODO: go through options of docker, docker-compose and sudo
type Options struct {
	Arguments   []string
	Filename    string
	Interactive bool
	Remove      bool
	NoCache     bool
	Pull        bool
	Build       bool
	Workdir     string
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
	if options.Workdir != "" {
		config.WorkingDir = options.Workdir
	}
}
