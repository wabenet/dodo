package command

import (
	"github.com/oclaussen/dodo/pkg/config"
	"github.com/oclaussen/dodo/pkg/container"
	"github.com/oclaussen/dodo/pkg/stage"
	"github.com/spf13/cobra"
)

func NewDaemonCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "daemon",
		Short:            "run backdrops in daemon mode",
		TraverseChildren: true,
		SilenceUsage:     true,
	}

	cmd.AddCommand(NewDaemonStartCommand())
	cmd.AddCommand(NewDaemonStopCommand())
	cmd.AddCommand(NewDaemonRestartCommand())
	return cmd
}

func NewDaemonStartCommand() *cobra.Command {
	var opts options
	cmd := &cobra.Command{
		Use:   "start",
		Short: "run a backdrop in daemon mode",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configureLogging()

			conf, err := config.LoadBackdrop(args[0])
			if err != nil {
				return err
			}

			optsConfig, err := opts.createConfig([]string{})
			if err != nil {
				return err
			}

			conf.Merge(optsConfig)

			return withStage(conf.Stage, func(s stage.Stage) error {
				c, err := container.NewContainer(conf, s, config.LoadAuthConfig(), true)
				if err != nil {
					return err
				}

				return c.Run()
			})
		},
	}
	opts.createFlags(cmd)

	return cmd
}

func NewDaemonStopCommand() *cobra.Command {
	var opts options
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "stop a daemon backdrop",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configureLogging()

			conf, err := config.LoadBackdrop(args[0])
			if err != nil {
				return err
			}

			optsConfig, err := opts.createConfig([]string{})
			if err != nil {
				return err
			}

			conf.Merge(optsConfig)

			return withStage(conf.Stage, func(s stage.Stage) error {
				c, err := container.NewContainer(conf, s, config.LoadAuthConfig(), true)
				if err != nil {
					return err
				}

				return c.Stop()
			})
		},
	}
	opts.createFlags(cmd)

	return cmd
}

func NewDaemonRestartCommand() *cobra.Command {
	var opts options
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "restart a daemon backdrop",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configureLogging()

			conf, err := config.LoadBackdrop(args[0])
			if err != nil {
				return err
			}

			optsConfig, err := opts.createConfig([]string{})
			if err != nil {
				return err
			}

			conf.Merge(optsConfig)

			return withStage(conf.Stage, func(s stage.Stage) error {
				c, err := container.NewContainer(conf, s, config.LoadAuthConfig(), true)
				if err != nil {
					return err
				}

				if err := c.Stop(); err != nil {
					return err
				}

				return c.Run()
			})
		},
	}
	opts.createFlags(cmd)

	return cmd
}
