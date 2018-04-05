package main

import (
	"github.com/oclaussen/dodo/context"
	"github.com/oclaussen/dodo/options"
	"github.com/spf13/cobra"
)

// TODO: no error message when bind mount fails

// TODO: do we need logging?
// TODO: tests, linting
func main() {
	var opts options.Options
	cmd := &cobra.Command{
		Use:              "dodo [OPTIONS] CONTEXT [CMD...]",
		Short:            "Run commands in a Docker context",
		Version:          "0.0.1", // TODO: fix help/version/errors
		TraverseChildren: true,
		Args:             cobra.MinimumNArgs(1),
		RunE:             func(cmd *cobra.Command, args []string) error {
			context := context.NewContext(args[0], &opts)
			return context.Run(args[1:])
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Filename, "file", "f", "", "Specify a dodo configuration file")
	flags.BoolVarP(&opts.Remove, "rm", "", false, "Automatically remove the container when it exits")
	flags.BoolVarP(&opts.NoCache, "no-cache", "", false, "Do not use cache when building the image")
	flags.BoolVarP(&opts.Pull, "pull", "", false, "Always attempt to pull a newer version of the image")
	flags.BoolVarP(&opts.Build, "build", "", false, "Always build an image, even if already exists")
	flags.SetInterspersed(false)

	cmd.Execute()
}
