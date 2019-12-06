package dockermachine

import (
	"encoding/json"
	"os/user"
	"path/filepath"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/check"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/mitchellh/mapstructure"
	"github.com/oclaussen/dodo/pkg/stage"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Stage struct {
	name    string
	Options *Options
	basedir string
	api     libmachine.API
}

type Options struct {
	Driver string
}

func (s *Stage) Initialize(name string, config *types.Stage) (bool, error) {
	s.Options = &Options{}
	if err := mapstructure.Decode(config.Options, s.Options); err != nil {
		return false, err
	}

	if len(s.Options.Driver) == 0 {
		s.Options.Driver = "virtualbox"
	}

	user, err := user.Current()
	if err != nil || user.HomeDir == "" {
		return false, errors.New("could not determine home directory")
	}

	s.name = name
	s.basedir = filepath.Join(user.HomeDir, ".docker", "machine")
	s.api = libmachine.NewClient(s.basedir, filepath.Join(s.basedir, "certs"))
	return true, nil
}

func (s *Stage) Create() error {
	driverConfig, _ := json.Marshal(&drivers.BaseDriver{
		MachineName: s.name,
		StorePath:   s.basedir,
	})

	target, err := s.api.NewHost(s.Options.Driver, driverConfig)
	if err != nil {
		return errors.Wrap(err, "could not create stage")
	}

	target.HostOptions = &host.Options{
		AuthOptions: &auth.Options{
			StorePath:        filepath.Join(s.basedir, "machines", s.name),
			CertDir:          filepath.Join(s.basedir, "certs"),
			CaCertPath:       filepath.Join(s.basedir, "certs", "ca.pem"),
			CaPrivateKeyPath: filepath.Join(s.basedir, "certs", "ca-key.pem"),
			ClientCertPath:   filepath.Join(s.basedir, "certs", "cert.pem"),
			ClientKeyPath:    filepath.Join(s.basedir, "certs", "key.pem"),
			ServerCertPath:   filepath.Join(s.basedir, "machines", s.name, "server.pem"),
			ServerKeyPath:    filepath.Join(s.basedir, "machines", s.name, "server-key.pem"),
		},
		EngineOptions: &engine.Options{
			TLSVerify: true,
		},
		SwarmOptions: &swarm.Options{},
	}

	if err = s.api.Create(target); err != nil {
		return errors.Wrap(err, "could not create stage")
	}

	if err := s.api.Save(target); err != nil {
		return errors.Wrap(err, "could not store stage")
	}

	log.WithFields(log.Fields{"name": s.name}).Info("stage is now up")
	return nil
}

func (s *Stage) Start() error {
	target, err := s.api.Load(s.name)
	if err != nil {
		return errors.Wrap(err, "could not load stage")
	}

	if err := target.Start(); err != nil {
		return errors.Wrap(err, "could not start stage")
	}

	if err := s.api.Save(target); err != nil {
		return errors.Wrap(err, "could not store stage")
	}

	log.WithFields(log.Fields{"name": s.name}).Info("stage is now up")
	return nil
}

func (s *Stage) Stop() error {
	target, err := s.api.Load(s.name)
	if err != nil {
		return errors.Wrap(err, "could not load stage")
	}

	if err := target.Stop(); err != nil {
		return errors.Wrap(err, "could not pause stage")
	}

	if err := s.api.Save(target); err != nil {
		return errors.Wrap(err, "could not store stage")
	}

	log.WithFields(log.Fields{"name": s.name}).Info("paused stage")
	return nil
}

func (s *Stage) Remove(force bool) error {
	if exist, err := s.Exist(); err == nil && !exist && !force {
		log.WithFields(log.Fields{"name": s.name}).Info("stage is not up")
		return nil
	}

	target, err := s.api.Load(s.name)
	if err != nil {
		return errors.Wrap(err, "could not load stage")
	}

	if err := target.Driver.Remove(); err != nil && !force {
		return errors.Wrap(err, "could not remove remote stage")
	}

	if err := s.api.Remove(s.name); err != nil && !force {
		return errors.Wrap(err, "could not remove local stage")
	}

	log.WithFields(log.Fields{"name": s.name}).Info("removed stage")
	return nil
}

func (s *Stage) Exist() (bool, error) {
	return s.api.Exists(s.name)
}

func (s *Stage) Available() (bool, error) {
	target, err := s.api.Load(s.name)
	if err != nil {
		return false, errors.Wrap(err, "could not load stage")
	}

	current, err := target.Driver.GetState()
	if err != nil {
		return false, errors.Wrap(err, "could not get state")
	}

	return current == state.Running, nil
}

func (s *Stage) GetSSHOptions() (*stage.SSHOptions, error) {
	target, err := s.api.Load(s.name)
	if err != nil {
		return nil, errors.Wrap(err, "could not load stage")
	}

	hostname, err := target.Driver.GetSSHHostname()
	if err != nil {
		return nil, errors.Wrap(err, "could not get SSH hostname")
	}

	port, err := target.Driver.GetSSHPort()
	if err != nil {
		return nil, errors.Wrap(err, "could not get SSH port")
	}

	return &stage.SSHOptions{
		Hostname:       hostname,
		Port:           port,
		Username:       target.Driver.GetSSHUsername(),
		PrivateKeyFile: target.Driver.GetSSHKeyPath(),
	}, nil
}

func (s *Stage) GetDockerOptions() (*stage.DockerOptions, error) {
	target, err := s.api.Load(s.name)
	if err != nil {
		return nil, errors.Wrap(err, "could not load stage")
	}

	dockerHost, _, err := check.DefaultConnChecker.Check(target, false)
	if err != nil {
		return nil, errors.Wrap(err, "could not check TLS connection")
	}

	return &stage.DockerOptions{
		Host:     dockerHost,
		CAFile:   filepath.Join(s.basedir, "machines", s.name, "ca.pem"),
		CertFile: filepath.Join(s.basedir, "machines", s.name, "cert.pem"),
		KeyFile:  filepath.Join(s.basedir, "machines", s.name, "key.pem"),
	}, nil
}
