package dodo

import (
	"errors"

	"github.com/wabenet/dodo-core/pkg/plugin"
	"github.com/wabenet/dodo-core/pkg/plugin/command"
)

const (
	ExitCodeInternalError = 1
	DefaultCommand        = "run"
)

var (
	ErrInvalidConfiguration = errors.New("invalid configuration")
)

func ExecuteDodoMain(m plugin.Manager) int {
	cmd := New(m, DefaultCommand).GetCobraCommand()

	if err := cmd.Execute(); err != nil {
		return ExitCodeInternalError
	}

	return command.GetExitCode(cmd)
}
