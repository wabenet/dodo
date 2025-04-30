package build

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wabenet/dodo-core/pkg/plugin"
	"github.com/wabenet/dodo-core/pkg/plugin/builder"
	"github.com/wabenet/dodo-core/pkg/plugin/configuration"
	"github.com/wabenet/dodo/pkg/core"
)

type options struct {
	noCache      bool
	forceRebuild bool
	forcePull    bool
	runtime      string
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
			backdrop, err := opts.createConfig(args[0])
			if err != nil {
				return fmt.Errorf("could not build backdrop image: %w", err)
			}

			if _, err := core.BuildByName(m, args[0], backdrop); err != nil {
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
	flags.StringVarP(
		&opts.runtime, "runtime", "r", "",
		"select runtime plugin")

	return &Command{cmd: cmd}
}

func (opts *options) createConfig(name string) (configuration.Backdrop, error) {
	c := configuration.Backdrop{
		Builder: opts.runtime,
		BuildConfig: builder.BuildConfig{
			ImageName:    name,
			NoCache:      opts.noCache,
			ForceRebuild: opts.forceRebuild,
			ForcePull:    opts.forcePull,
		},
	}

	return c, nil
}
