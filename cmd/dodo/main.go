package main

import (
	"os"

	"github.com/oclaussen/dodo/pkg/command"
	"github.com/oclaussen/dodo/pkg/container"
)

func main() {
	cmd := command.NewCommand()
	if err := cmd.Execute(); err != nil {
		if err, ok := err.(*container.ScriptError); ok {
			os.Exit(err.ExitCode)
		} else {
			os.Exit(1)
		}
	}
}
