package designer

import (
	"fmt"
	"net"
	"time"

	"github.com/oclaussen/go-gimme/ssl"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func Provision(config *Config) (*ProvisionResult, error) {
	log.Info("replace insecure SSH key")
	if err := ConfigureSSHKeys(config); err != nil {
		return nil, err
	}

	log.Info("configure host network...")
	ip, err := ConfigureNetwork(Network{Device: "eth1"})
	if err != nil {
		return nil, err
	}

	log.Info("set hostname...")
	if err := ConfigureHostname(config.Hostname); err != nil {
		return nil, err
	}

	log.Info("installing docker...")
	if err := InstallDocker(); err != nil {
		return nil, err
	}

	certs, _, err := ssl.GimmeCertificates(&ssl.Options{
		Org:   fmt.Sprintf("dodo.%s", config.Hostname),
		Hosts: []string{ip, "localhost"},
	})
	if err != nil {
		return nil, err
	}

	log.Info("configuring docker...")
	if err := ConfigureDocker(&DockerConfig{
		CA:          certs.CA,
		ServerCert:  certs.ServerCert,
		ServerKey:   certs.ServerKey,
		Environment: config.Environment,
		Arguments:   config.DockerArgs,
	}); err != nil {
		return nil, err
	}

	log.Info("starting docker...")
	if err := RestartDocker(); err != nil {
		return nil, err
	}

	result := &ProvisionResult{
		IPAddress:  ip,
		CA:         string(certs.CA),
		ClientCert: string(certs.ClientCert),
		ClientKey:  string(certs.ClientKey),
	}

	for attempts := 0; attempts < 60; attempts++ {
		if conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", dockerPort)); err == nil {
			conn.Close()
			return result, nil
		}
		time.Sleep(5 * time.Second)
	}

	return nil, errors.New("docker did not start successfully")
}
