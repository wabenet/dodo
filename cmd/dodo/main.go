package main

import (
	"os"

	"github.com/oclaussen/dodo/pkg/command"
	log "github.com/sirupsen/logrus"
)

// TODO: add a readme and license
// TODO: reduce linter to 80 chars
// TODO: provide some context to all of the error messages
// (both in logging and return values)

// TODO automatically guess image name based on backdrop name
// TODO automatically set image name for caching

func main() {
	cmd := command.NewCommand()
	if err := cmd.Execute(); err != nil {
		// TODO: printing the error is done by both cobra and logrus
		log.Error(err)
		os.Exit(1)
	}
}
