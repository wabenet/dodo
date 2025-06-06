package dodo

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/wabenet/dodo-core/pkg/plugin"
	"github.com/wabenet/dodo-core/pkg/plugin/command"
	"go.szostok.io/version/extension"
)

func New(m plugin.Manager, defaultCmd string) *Command {
	cmd := &cobra.Command{
		Use:                Name,
		SilenceUsage:       true,
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				if self, err := os.Executable(); err == nil {
					return runProxy(cmd, self, []string{defaultCmd})
				}

				return plugin.NewNotFoundError(command.Type, defaultCmd)
			}

			if path, err := exec.LookPath(fmt.Sprintf("dodo-%s", args[0])); err == nil {
				return runProxy(cmd, path, args[1:])
			}

			path := plugin.PathByName(args[0])
			if stat, err := os.Stat(path); err == nil && stat.Mode().Perm()&0111 != 0 {
				return runProxy(cmd, path, args[1:])
			}

			if self, err := os.Executable(); err == nil {
				return runProxy(cmd, self, append([]string{defaultCmd}, args...))
			}

			return plugin.NewNotFoundError(command.Type, defaultCmd)
		},
	}

	cmd.AddCommand(extension.NewVersionCobraCmd())

	for _, p := range m.GetPlugins(command.Type.String()) {
		cmd.AddCommand(p.(command.Command).GetCobraCommand())
	}

	return &Command{cmd: cmd}
}

func runProxy(cmd *cobra.Command, executable string, args []string) error {
	run := exec.Command(executable, args...)

	run.Stdin = os.Stdin
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr

	if err := run.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			command.SetExitCode(cmd, exit.ExitCode())
		} else {
			return fmt.Errorf("error executing external command: %w", err)
		}
	}

	return nil
}
