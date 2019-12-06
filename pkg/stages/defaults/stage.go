package defaults

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/go-plugin"
	"github.com/oclaussen/dodo/pkg/stage"
	"github.com/oclaussen/dodo/pkg/stages/environment"
	"github.com/oclaussen/dodo/pkg/stages/generic"
	"github.com/oclaussen/dodo/pkg/stages/grpc"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/oclaussen/go-gimme/configfiles"
	"github.com/pkg/errors"
)

const DefaultStageName = "environment"

var builtinStages = map[string]stage.Stage{
	"environment": &environment.Stage{},
	"generic":     &generic.Stage{},
}

// TODO: sort out when and how to cleanup the plugin process properly

func Load(name string, conf *types.Stage) (stage.Stage, func(), error) {
	if conf == nil {
		conf = &types.Stage{Type: DefaultStageName}
	}
	if stage, ok := builtinStages[conf.Type]; ok {
		if success, err := stage.Initialize(name, conf); err != nil || !success {
			return nil, func() {}, errors.Wrap(err, "initialization failed")
		}
		return stage, func() {}, nil
	}

	path, err := findPluginExecutable(conf.Type)
	if err != nil {
		return nil, func() {}, err
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  grpc.HandshakeConfig(conf.Type),
		Plugins:          grpc.PluginMap,
		Cmd:              exec.Command(path),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
		Logger:           stage.NewPluginLogger(),
	})

	c, err := client.Client()
	if err != nil {
		return nil, client.Kill, err
	}
	raw, err := c.Dispense("stage")
	if err != nil {
		return nil, client.Kill, err
	}

	stage := raw.(stage.Stage)
	if success, err := stage.Initialize(name, conf); err != nil || !success {
		return nil, client.Kill, errors.Wrap(err, "initialization failed")
	}

	return stage, client.Kill, nil
}

func findPluginExecutable(name string) (string, error) {
	directories, err := configfiles.GimmeConfigDirectories(&configfiles.Options{
		Name:                      "dodo",
		IncludeWorkingDirectories: true,
	})
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("plugin-%s_%s_%s", name, runtime.GOOS, runtime.GOARCH)
	for _, dir := range directories {
		path := filepath.Join(dir, ".dodo", "plugins", filename)
		if stat, err := os.Stat(path); err == nil && stat.Mode().Perm()&0111 != 0 {
			return path, nil
		}
	}

	return "", errors.New("could not find a suitable plugin for the stage anywhere")
}
