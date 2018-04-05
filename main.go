package main

import (
	"github.com/oclaussen/dodo/context"
	"github.com/oclaussen/dodo/options"
	"github.com/spf13/cobra"
)

// TODO: do we need logging?
// TODO: tests, linting
func main() {
	var opts options.Options
	cmd := &cobra.Command{
		Use:              "dodo [OPTIONS] CONTEXT [CMD...]",
		Short:            "blub", // TODO: description
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
	flags.SetInterspersed(false)

	cmd.Execute()
}
