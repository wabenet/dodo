package main

import (
	"fmt"
	"os"

	"github.com/oclaussen/dodo/pkg/command"
)

func main() {
	cmd := command.NewCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
