package main

import (
	"os"

	"github.com/oclaussen/dodo/pkg/command"
	log "github.com/sirupsen/logrus"
)

// TODO: add a readme and license

// TODO automatically set image name for caching

func main() {
	cmd := command.NewCommand()
	if err := cmd.Execute(); err != nil {
		log.Debug(err)
		os.Exit(1)
	}
}
