package dodo

import (
	"github.com/spf13/cobra"
	api "github.com/wabenet/dodo-core/api/v1alpha4"
	"github.com/wabenet/dodo-core/pkg/plugin"
	"github.com/wabenet/dodo-core/pkg/plugin/command"
)

const Name = "dodo"

var _ command.Command = &Command{}

type Command struct {
	cmd *cobra.Command
}

func (p *Command) Type() plugin.Type {
	return command.Type
}

func (p *Command) PluginInfo() *api.PluginInfo {
	return &api.PluginInfo{
		Name: &api.PluginName{Name: Name, Type: command.Type.String()},
	}
}

func (*Command) Init() (plugin.PluginConfig, error) {
	return map[string]string{}, nil
}

func (*Command) Cleanup() {}

func (p *Command) GetCobraCommand() *cobra.Command {
	return p.cmd
}
