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
	opts := options.NewOptions()
	cmd := &cobra.Command{
		Use:              "dodo [OPTIONS] CONTEXT [CMD...]",
		Short:            "Run commands in a Docker context",
		Version:          "0.0.1", // TODO: fix help/version/errors
		TraverseChildren: true,
		Args:             cobra.MinimumNArgs(1),
		RunE:             func(cmd *cobra.Command, args []string) error {
			context := context.NewContext(args[0], opts)
			return context.Run(args[1:])
		},
	}
	opts.CreateFlags(cmd)
	cmd.Execute()
}
