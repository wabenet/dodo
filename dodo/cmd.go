package dodo

import (
	"github.com/oclaussen/dodo/command"
	"github.com/oclaussen/dodo/config"
	"github.com/spf13/cobra"
)

type dodoOptions struct {
	file        string
	command     string
	arguments   []string
}

// TODO: allow all docker flags
func NewCommand() *cobra.Command {
	var opts dodoOptions

	cmd := &cobra.Command{
		Use:              "dodo [OPTIONS] CMD [ARG...]",
		Short:            "blub", // TODO: description
		Version:          "0.0.1", // TODO: fix help/version/errors
		TraverseChildren: true,
		Args:             cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.command = args[0]
			opts.arguments = args[1:]
			return runDodo(opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.file, "file", "f", "dodo.yaml", "Specify an alternate dodo file")
	flags.SetInterspersed(false)

	return cmd
}

func runDodo(opts dodoOptions) error {
	config, err := config.Load(opts.file)
	if err != nil {
	  return err
	}

	command := command.NewCommand(config.Commands[opts.command])

	err = command.Run()
	if err != nil {
		return err
	}

	return nil
}
