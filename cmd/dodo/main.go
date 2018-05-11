package main

import (
	"os"

	"github.com/oclaussen/dodo/pkg/command"
	log "github.com/sirupsen/logrus"
)

func main() {
	cmd := command.NewCommand()
	if err := cmd.Execute(); err != nil {
		log.Debug(err)
		os.Exit(1)
	}
}
