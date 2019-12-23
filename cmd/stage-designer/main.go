package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/oclaussen/dodo/pkg/stagedesigner"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(new(log.JSONFormatter))

	configFile, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	config, err := stagedesigner.DecodeConfig(configFile)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	result, err := stagedesigner.Provision(config)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	output, err := stagedesigner.EncodeResult(result)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	fmt.Print(string(output))
}
