package main

import (
	plugin1 "github.com/dodo-cli/dodo-build/plugin"
	plugin2 "github.com/dodo-cli/dodo-config/plugin"
	appconfig "github.com/dodo-cli/dodo-core/pkg/appconfig"
	plugin "github.com/dodo-cli/dodo-core/pkg/plugin"
	command "github.com/dodo-cli/dodo-core/pkg/plugin/command"
	configuration "github.com/dodo-cli/dodo-core/pkg/plugin/configuration"
	runtime "github.com/dodo-cli/dodo-core/pkg/plugin/runtime"
	proxycmd "github.com/dodo-cli/dodo-core/pkg/proxycmd"
	run "github.com/dodo-cli/dodo-core/pkg/run"
	plugin3 "github.com/dodo-cli/dodo-docker/plugin"
	gohclog "github.com/hashicorp/go-hclog"
	"os"
)

func main() {
	gohclog.SetDefault(gohclog.New(appconfig.GetLoggerOptions()))
	os.Exit(execute())
}
func execute() int {
	includePlugins()
	plugin.LoadPlugins(command.Type, configuration.Type, runtime.Type)
	defer plugin.UnloadPlugins()
	return proxycmd.Execute("run")
}
func includePlugins() {
	plugin.IncludePlugins(&run.Command{})
	plugin1.IncludeMe()
	plugin2.IncludeMe()
	plugin3.IncludeMe()
}
