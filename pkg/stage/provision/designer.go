// +build designer

package main

import (
	"fmt"
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
	result, err := designer.Provision(config)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	output, err := designer.EncodeResult(result)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	fmt.Print(string(output))
}
