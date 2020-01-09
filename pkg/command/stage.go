package command

import (
	"github.com/oclaussen/dodo/pkg/stage"
	"github.com/oclaussen/go-gimme/ssh"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewStageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "stage",
		Short:            "Manage stages",
		TraverseChildren: true,
		SilenceUsage:     true,
	}

	cmd.AddCommand(NewUpCommand())
	cmd.AddCommand(NewDownCommand())
	cmd.AddCommand(NewSSHCommand())
	return cmd
}

func NewUpCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Create or start a stage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configureLogging()
			return withStage(args[0], func(s stage.Stage) error {
				exist, err := s.Exist()
				if err != nil {
					return err
				}

				if !exist {
					return s.Create()
				}
				return s.Start()
			})
		},
	}
}

type downOptions struct {
	remove  bool
	volumes bool
	force   bool
}

func NewDownCommand() *cobra.Command {
	var opts downOptions
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Remove or pause a stage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configureLogging()
			return withStage(args[0], func(s stage.Stage) error {
				if opts.remove {
					return s.Remove(opts.force, opts.volumes)
				}
				return s.Stop()
			})
		},
	}
	flags := cmd.Flags()
	flags.BoolVarP(&opts.remove, "rm", "", false, "remove the stage instead of pausing")
	flags.BoolVarP(&opts.volumes, "volumes", "", false, "when used with '--rm', also delete persistent volumes")
	flags.BoolVarP(&opts.force, "force", "f", false, "when used with '--rm', don't stop on errors")
	return cmd
}

func NewSSHCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ssh",
		Short: "login to the stage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configureLogging()
			return withStage(args[0], func(s stage.Stage) error {
				available, err := s.Available()
				if err != nil {
					return err
				}

				if !available {
					return errors.New("stage is not up")
				}

				opts, err := s.GetSSHOptions()
				if err != nil {
					return err
				}

				return ssh.GimmeShell(&ssh.Options{
					Host:              opts.Hostname,
					Port:              opts.Port,
					User:              opts.Username,
					IdentityFileGlobs: []string{opts.PrivateKeyFile},
					NonInteractive:    true,
				})
			})
		},
	}
}
