package main

import (
	"io/ioutil"
	"os"

	"github.com/dave/jennifer/jen"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Plugins []*Plugin `yaml:"plugins"`
}

type Plugin struct {
	Import string `yaml:"import"`
}

func main() {
	bytes, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(bytes, cfg); err != nil {
		panic(err)
	}

	f := jen.NewFile("main")

	f.Func().Id("main").Params().Block(
		// Configure Logging
		jen.Qual("github.com/hashicorp/go-hclog", "SetDefault").Call(
			jen.Qual("github.com/hashicorp/go-hclog", "New").Call(
				jen.Qual("github.com/dodo-cli/dodo-core/pkg/appconfig", "GetLoggerOptions").Call(),
			),
		),

		// "Actual main" to keep os.Exit from interfering with defers
		jen.Qual("os", "Exit").Call(jen.Id("execute").Call()),
	)

	f.Func().Id("execute").Params().Int().Block(
		jen.Id("includePlugins").Call(),
		jen.Qual("github.com/dodo-cli/dodo-core/pkg/plugin", "LoadPlugins").Call(),
		jen.Defer().Qual("github.com/dodo-cli/dodo-core/pkg/plugin", "UnloadPlugins").Call(),
		jen.Return(jen.Qual("github.com/dodo-cli/dodo-core/pkg/proxycmd", "Execute").Call(jen.Lit("run"))),
	)

	f.Func().Id("includePlugins").Params().BlockFunc(func(g *jen.Group) {
		for _, p := range cfg.Plugins {
			g.Qual(p.Import, "IncludeMe").Call()
		}
	})

	f.Save("main.go")
}
