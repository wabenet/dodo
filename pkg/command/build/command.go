package build

import (
	"fmt"

	api "github.com/dodo-cli/dodo-core/api/v1alpha2"
	"github.com/dodo-cli/dodo-core/pkg/plugin"
	"github.com/dodo-cli/dodo/pkg/core"
	"github.com/spf13/cobra"
)

type options struct {
	noCache      bool
	forceRebuild bool
	forcePull    bool
}

func New(m plugin.Manager) *Command {
	var opts options

	cmd := &cobra.Command{
		Use:                   Name,
		Short:                 "Build all required images for backdrop without running it",
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Args:                  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := &api.BuildInfo{
				ImageName:    args[0],
				NoCache:      opts.noCache,
				ForceRebuild: opts.forceRebuild,
				ForcePull:    opts.forcePull,
			}

			if _, err := core.BuildByName(m, config); err != nil {
				return fmt.Errorf("could not build backdrop image: %w", err)
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)

	flags.BoolVar(
		&opts.noCache, "no-cache", false,
		"do not use cache when building the image")
	flags.BoolVarP(
		&opts.forceRebuild, "force", "f", false,
		"always rebuild all dependencies, even when they already exist")
	flags.BoolVar(
		&opts.forcePull, "pull", false,
		"always attempt to pull base images")

	return &Command{cmd: cmd}
}
