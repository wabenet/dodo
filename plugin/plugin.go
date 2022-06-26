package plugin

import (
	"github.com/wabenet/dodo-core/pkg/plugin"
	"github.com/wabenet/dodo/pkg/command/build"
	"github.com/wabenet/dodo/pkg/command/run"
)

func IncludeMe(m plugin.Manager) {
	m.IncludePlugins(run.New(m), build.New(m))
}
