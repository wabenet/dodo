package stage

import (
	"fmt"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/docker/docker/client"
	"github.com/hashicorp/go-plugin"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/pkg/errors"
)

var BuiltInStages = map[string]Stage{
	DefaultStageName: &DefaultStage{},
}

type Stage interface {
	Initialize(map[string]string) (bool, error)
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

// TODO: make machine dir configurable and default somewhere not docker-machine
// TODO: sort out when and how to cleanup the plugin process properly

func Load(name string, config *types.Stage) (Stage, func(), error) {
	stateDir := filepath.FromSlash("/tmp/docker/machine")

	if user, err := user.Current(); err == nil && user.HomeDir != "" {
		stateDir = filepath.Join(user.HomeDir, ".docker", "machine")
	}

	if stage, ok := BuiltInStages[config.Type]; ok {
		return stage, func() {}, nil
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  HandshakeConfig(config.Type),
		Plugins:          PluginMap,
		Cmd:              exec.Command(fmt.Sprintf("./%s", config.Type)),
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
	success, err := stage.Initialize(map[string]string{
		"vmName":      name,
		"storagePath": filepath.Join(stateDir, "machines", name),
		"cachePath":   filepath.Join(stateDir, "cache"),
	})
	if err != nil || !success {
		return nil, client.Kill, errors.Wrap(err, "initialization failed")
	}

	return stage, client.Kill, nil
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
