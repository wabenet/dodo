package stage

import (
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/docker/docker/client"
	"github.com/hashicorp/go-plugin"
	"github.com/oclaussen/dodo/pkg/stage/provider"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/pkg/errors"
)

// TODO: make machine dir configurable and default somewhere not docker-machine

type Stage struct {
	name     string
	config   *types.Stage
	stateDir string
	client   *plugin.Client
	provider provider.Provider
}

func LoadStage(name string, config *types.Stage) (*Stage, error) {
	stage := &Stage{
		name:     name,
		config:   config,
		stateDir: filepath.Join(home(), ".docker", "machine"),
	}

	if prov, ok := provider.BuiltInProviders[config.Type]; ok {
		stage.provider = prov
		return stage, nil
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  provider.HandshakeConfig("virtualbox"),
		Plugins:          provider.PluginMap,
		Cmd:              exec.Command("./virtualbox"),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
		Logger:           provider.NewPluginLogger(),
	})
	stage.client = client

	c, err := client.Client()
	if err != nil {
		return nil, err
	}
	raw, err := c.Dispense("provider")
	if err != nil {
		return nil, err
	}

	stage.provider = raw.(provider.Provider)
	success, err := stage.provider.Initialize(map[string]string{
		"vmName":      name,
		"storagePath": stage.hostDir(),
		"cachePath":   filepath.Join(stage.stateDir, "cache"),
	})
	if err != nil || !success {
		return nil, errors.Wrap(err, "initialization failed")
	}

	return stage, nil
}

func (stage *Stage) Save() {
	if stage.client != nil {
		stage.client.Kill()
	}
}

func (stage *Stage) Up() error {
	exist, err := stage.provider.Exist()
	if err != nil {
		return err
	}
	if exist {
		return stage.provider.Start()
	} else {
		return stage.provider.Create()
	}
}

func (stage *Stage) Down(remove bool, force bool) error {
	if remove {
		return stage.provider.Remove(force)
	} else {
		return stage.provider.Stop()
	}
}

func (stage *Stage) GetDockerClient() (*client.Client, error) {
	available, err := stage.provider.Available()
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, errors.New("stage is not up")
	}
	opts, err := stage.provider.GetDockerOptions()
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

func (stage *Stage) hostDir() string {
	return filepath.Join(stage.stateDir, "machines", stage.name)
}

func home() string {
	user, err := user.Current()
	if err != nil {
		return filepath.FromSlash("/")
	}
	if user.HomeDir == "" {
		return filepath.FromSlash("/")
	}
	return user.HomeDir
}
