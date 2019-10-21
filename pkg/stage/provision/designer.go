// +build designer

package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/oclaussen/dodo/pkg/stage/designer"
)

func main() {
	configFile, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	config, err := designer.DecodeConfig(configFile)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	if err := designer.Provision(config); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
