package command

import (
	"github.com/oclaussen/dodo/pkg/config"
	"github.com/oclaussen/dodo/pkg/stages/defaults"
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
			conf, err := config.LoadStage(args[0])
			if err != nil {
				return err
			}
			s, cleanup, err := defaults.Load(args[0], conf)
			defer cleanup()
			if err != nil {
				return err
			}

			exist, err := s.Exist()
			if err != nil {
				return err
			}

			if !exist {
				return s.Create()
			}
			return s.Start()
		},
	}
}

type downOptions struct {
	remove bool
	force  bool
}

func NewDownCommand() *cobra.Command {
	var opts downOptions
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Remove or pause a stage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configureLogging()
			// TODO: do we actually need the config?
			conf, err := config.LoadStage(args[0])
			if err != nil {
				return err
			}
			s, cleanup, err := defaults.Load(args[0], conf)
			defer cleanup()
			if err != nil {
				return err
			}

			if opts.remove {
				return s.Remove(opts.force)
			}
			return s.Stop()
		},
	}
	flags := cmd.Flags()
	flags.BoolVarP(&opts.remove, "rm", "", false, "remove the stage instead of pausing")
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
			// TODO: do we actually need the config?
			conf, err := config.LoadStage(args[0])
			if err != nil {
				return err
			}
			s, cleanup, err := defaults.Load(args[0], conf)
			defer cleanup()
			if err != nil {
				return err
			}

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
		},
	}
}
