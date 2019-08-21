// +build designer

package main

import (
	"log"
	"os"

	"github.com/oclaussen/dodo/pkg/stage/designer"
)

func main() {
	config, err := designer.DecodeConfig([]byte(os.Args[1]))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	if err := designer.Provision(config); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
