package run

import (
	"github.com/spf13/cobra"
	"github.com/wabenet/dodo-core/pkg/plugin"
	"github.com/wabenet/dodo-core/pkg/plugin/command"
)

const Name = "run"

var _ command.Command = &Command{}

type Command struct {
	cmd *cobra.Command
}

func (p *Command) Type() plugin.Type {
	return command.Type
}

func (p *Command) Metadata() plugin.Metadata {
	return plugin.NewMetadata(command.Type, Name)
}

func (*Command) Init() (plugin.Config, error) {
	return map[string]string{}, nil
}

func (*Command) Cleanup() {}

func (p *Command) GetCobraCommand() *cobra.Command {
	return p.cmd
}
