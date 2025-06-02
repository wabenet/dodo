package main

import (
	dodobuildkit "github.com/wabenet/dodo-buildkit"
	dodoconfig "github.com/wabenet/dodo-config"
	dodocore "github.com/wabenet/dodo-core"
	plugin "github.com/wabenet/dodo-core/pkg/plugin"
	dododocker "github.com/wabenet/dodo-docker"
	dodo "github.com/wabenet/dodo/pkg/command/dodo"
	plugin1 "github.com/wabenet/dodo/plugin"
	"os"
)

func main() {
	os.Exit(execute())
}
func execute() int {
	m := plugin.Init()
	includePlugins(m)
	m.LoadPlugins()
	defer m.UnloadPlugins()
	return dodo.ExecuteDodoMain(m)
}
func includePlugins(m plugin.Manager) {
	plugin1.IncludeMe(m)
	dodocore.IncludeMe(m)
	dodoconfig.IncludeMe(m)
	dodobuildkit.IncludeMe(m)
	dododocker.IncludeMe(m)
}
