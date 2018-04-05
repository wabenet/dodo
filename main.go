package main

import (
	"github.com/oclaussen/dodo/state"
	"github.com/oclaussen/dodo/options"
	"github.com/spf13/cobra"
)

// TODO: no error message when bind mount fails

// TODO: do we need logging?
// TODO: tests, linting
func main() {
	var opts options.Options
	cmd := &cobra.Command{
		Use:              "dodo [OPTIONS] NAME [CMD...]",
		Short:            "Run commands in a Docker context",
		Version:          "0.0.1", // TODO: fix help/version/errors
		TraverseChildren: true,
		Args:             cobra.MinimumNArgs(1),
		RunE:             func(cmd *cobra.Command, args []string) error {
			opts.Arguments = args[1:]
			state := state.NewState(args[0], &opts)
			return state.Run()
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Filename, "file", "f", "", "Specify a dodo configuration file")
	flags.BoolVarP(&opts.Interactive, "interactive", "i", false, "Run an interactive session")
	flags.BoolVarP(&opts.Remove, "rm", "", false, "Automatically remove the container when it exits")
	flags.BoolVarP(&opts.NoCache, "no-cache", "", false, "Do not use cache when building the image")
	flags.BoolVarP(&opts.Pull, "pull", "", false, "Always attempt to pull a newer version of the image")
	flags.BoolVarP(&opts.Build, "build", "", false, "Always build an image, even if already exists")
	flags.StringVarP(&opts.Workdir, "workdir", "w", "", "Working directory inside the container")
	flags.SetInterspersed(false)

	cmd.Execute()
}
