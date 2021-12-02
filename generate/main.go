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
		// "Actual main" to keep os.Exit from interfering with defers
		jen.Qual("os", "Exit").Call(jen.Id("execute").Call()),
	)

	f.Func().Id("execute").Params().Int().Block(
		jen.Id("m").Op(":=").Qual("github.com/dodo-cli/dodo-core/pkg/plugin", "Init").Call(),
		jen.Id("includePlugins").Call(jen.Id("m")),
		jen.Id("m").Dot("LoadPlugins").Call(),
		jen.Defer().Id("m").Dot("UnloadPlugins").Call(),
                jen.Return(jen.Qual("github.com/dodo-cli/dodo/pkg/command/dodo", "ExecuteDodoMain").Call(jen.Id("m"))),
	)

	f.Func().Id("includePlugins").Params(jen.Id("m").Qual("github.com/dodo-cli/dodo-core/pkg/plugin", "Manager")).BlockFunc(func(g *jen.Group) {
		for _, p := range cfg.Plugins {
			g.Qual(p.Import, "IncludeMe").Call(jen.Id("m"))
		}
	})

	f.Save("main.go")
}
