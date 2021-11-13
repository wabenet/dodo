package plugin

import (
	"github.com/dodo-cli/dodo-core/pkg/plugin"
	"github.com/dodo-cli/dodo/pkg/command/build"
	"github.com/dodo-cli/dodo/pkg/command/run"
)

func IncludeMe(m plugin.Manager) {
	m.IncludePlugins(run.New(m), build.New(m))
}
