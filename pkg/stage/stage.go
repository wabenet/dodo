package stage

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/docker/docker/client"
	"github.com/hashicorp/go-plugin"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/oclaussen/go-gimme/configfiles"
	"github.com/pkg/errors"
)

const (
	DefaultStageName  = "environment"
	DefaultAPIVersion = "1.39"
)

var BuiltInStages = map[string]Stage{
	"environment": &EnvStage{},
	"generic":     &GenericStage{},
}

type Stage interface {
	Initialize(string, map[string]string) (bool, error)
	Create() error
	Start() error
	Stop() error
	Remove(bool) error
	Exist() (bool, error)
	Available() (bool, error)
	GetSSHOptions() (*SSHOptions, error)
	GetDockerOptions() (*DockerOptions, error)
}

type SSHOptions struct {
	Hostname       string
	Port           int
	Username       string
	PrivateKeyFile string
}

type DockerOptions struct {
	Version  string
	Host     string
	CAFile   string
	CertFile string
	KeyFile  string
}

// TODO: sort out when and how to cleanup the plugin process properly

func Load(name string, config *types.Stage) (Stage, func(), error) {
	if stage, ok := BuiltInStages[config.Type]; ok {
		return stage, func() {}, nil
	}

	path, err := findPluginExecutable(config.Type)
	if err != nil {
		return nil, func() {}, err
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  HandshakeConfig(config.Type),
		Plugins:          PluginMap,
		Cmd:              exec.Command(path),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
		Logger:           NewPluginLogger(),
	})

	c, err := client.Client()
	if err != nil {
		return nil, client.Kill, err
	}
	raw, err := c.Dispense("stage")
	if err != nil {
		return nil, client.Kill, err
	}

	stage := raw.(Stage)
	if success, err := stage.Initialize(name, config.Options); err != nil || !success {
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

	filename := fmt.Sprintf("%s_%s_%s", name, runtime.GOOS, runtime.GOARCH)
	for _, dir := range directories {
		path := filepath.Join(dir, ".dodo", "plugins", filename)
		if stat, err := os.Stat(path); err == nil && stat.Mode().Perm()&0111 != 0 {
			return path, nil
		}
	}

	return "", errors.New("could not find a suitable plugin for the stage anywhere")

}

func GetDockerClient(stage Stage) (*client.Client, error) {
	available, err := stage.Available()
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, errors.New("stage is not up")
	}
	opts, err := stage.GetDockerOptions()
	if err != nil {
		return nil, err
	}
	mutators := []client.Opt{}
	if len(opts.Version) > 0 {
		mutators = append(mutators, client.WithVersion(opts.Version))
	}
	if len(opts.Host) > 0 {
		mutators = append(mutators, client.WithHost(opts.Host))
	}
	if len(opts.CAFile)+len(opts.CertFile)+len(opts.KeyFile) > 0 {
		mutators = append(mutators, client.WithTLSClientConfig(opts.CAFile, opts.CertFile, opts.KeyFile))
	}
	return client.NewClientWithOpts(mutators...)
}
