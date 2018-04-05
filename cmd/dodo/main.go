package main

import (
	"github.com/oclaussen/dodo/pkg/command"
)

func main() {
	cmd := command.NewCommand()
	cmd.Execute()
}
