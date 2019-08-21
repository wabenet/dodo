// +build ignore

package main

import (
	"net/http"

	"github.com/shurcool/vfsgen"
	log "github.com/sirupsen/logrus"
)

var fs http.FileSystem = http.Dir("./assets/")

func main() {
	err := vfsgen.Generate(fs, vfsgen.Options{
		PackageName:  "provision",
		VariableName: "Assets",
		BuildTags:    "!designer",
	})
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("could not generate assets")
	}
}
