package main

import (
	"os"

	"github.com/oclaussen/dodo/pkg/command"
	"github.com/oclaussen/dodo/pkg/types"
)

func main() {
	cmd := command.NewCommand()
	if err := cmd.Execute(); err != nil {
		if err, ok := err.(*types.ScriptError); ok {
			os.Exit(err.ExitCode)
		} else {
			os.Exit(1)
		}
	}
}
