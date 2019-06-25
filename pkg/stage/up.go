package stage

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/state"
	"github.com/oclaussen/go-gimme/ssl"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (stage *Stage) Up() error {
	if stage.exists {
		return stage.start()
	} else {
		return stage.create()
	}
}

func (stage *Stage) start() error {
	log.WithFields(log.Fields{"name": stage.name}).Info("starting stage...")

	currentState, err := stage.driver.GetState()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Debug("could not get machine state")
	}
	if currentState == state.Running {
		log.WithFields(log.Fields{"name": stage.name}).Info("stage is already up")
		return nil
	}

	if err := stage.driver.Start(); err != nil {
		return errors.Wrap(err, "could not start stage")
	}

	if err := stage.waitForDocker(); err != nil {
		return errors.Wrap(err, "docker did start successfully")
	}

	if err := stage.exportState(); err != nil {
		return errors.Wrap(err, "could not store stage")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("stage is now up")
	return nil

}

func (stage *Stage) create() error {
	driverOpts, err := stage.driverOptions()
	if err != nil {
		return err
	}
	if err = stage.driver.SetConfigFromFlags(driverOpts); err != nil {
		return errors.Wrap(err, "could not configure stage")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("running pre-create checks...")

	if err := stage.driver.PreCreateCheck(); err != nil {
		return err
	}

	if err := stage.exportState(); err != nil {
		return errors.Wrap(err, "could not save stage before creation")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("creating stage...")

	if err := stage.driver.Create(); err != nil {
		return errors.Wrap(err, "could not run driver")
	}

	if err := stage.exportState(); err != nil {
		return errors.Wrap(err, "could not save stage after creation")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("waiting for stage...")
	if err := stage.waitForState(state.Running); err != nil {
		return err
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("provisioning...")

	if err := stage.setHostname(); err != nil {
		return err
	}

	if err := stage.waitForDocker(); err != nil {
		return err
	}

	if err := stage.makeDockerOptionsDir(); err != nil {
		return err
	}

	ip, err := stage.driver.GetIP()
	if err != nil {
		return err
	}

	certs, files, err := ssl.GimmeCertificates(&ssl.Options{
		Org:          "dodo." + stage.name,
		Hosts:        []string{ip, "localhost"},
		WriteToFiles: &ssl.Files{Directory: stage.hostDir()},
	})
	if err != nil {
		return err
	}

	if err := stage.stopDocker(); err != nil {
		return err
	}

	if err := stage.deleteDockerLink(); err != nil {
		return err
	}

	log.Info("copying certs to the remote machine...")

	if err := stage.writeRemoteFile(files.CAFile, path.Join(dockerDir, "ca.pem")); err != nil {
		return err
	}
	if err := stage.writeRemoteFile(files.ServerCertFile, path.Join(dockerDir, "server.pem")); err != nil {
		return err
	}
	if err := stage.writeRemoteFile(files.ServerKeyFile, path.Join(dockerDir, "server-key.pem")); err != nil {
		return err
	}

	dockerURL, err := stage.driver.GetURL()
	if err != nil {
		return err
	}
	dockerPort, err := parseDockerPort(dockerURL)
	if err != nil {
		return err
	}

	if err := stage.writeDockerOptions(dockerPort); err != nil {
		return err
	}

	if err := stage.startDocker(); err != nil {
		return err
	}

	if err := stage.waitForDocker(); err != nil {
		return err
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, dockerPort), 5*time.Second)
	if err != nil {
		log.Warn(`
This machine has been allocated an IP address, but Docker Machine could not
reach it successfully.

SSH for the machine should still work, but connecting to exposed ports, such as
the Docker daemon port, may not work properly.

You may need to add the route manually, or use another related workaround.

This could be due to a VPN, proxy, or host file configuration issue.

You also might want to clear any VirtualBox host only interfaces you are not using.`)
	} else {
		conn.Close()
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("checking connection...")

	parsedURL, err := url.Parse(dockerURL)
	if err != nil {
		return errors.Wrap(err, "could not parse Docker URL")
	}

	if valid, err := validateCertificate(parsedURL.Host, certs); !valid || err != nil {
		return errors.Wrap(err, "invalid certificate")
	}

	if err := stage.exportState(); err != nil {
		return errors.Wrap(err, "could not store stage")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("stage is now up")
	return nil
}

func validateCertificate(addr string, certs *ssl.Certificates) (bool, error) {
	keyPair, err := tls.X509KeyPair(certs.CA, certs.CAKey)
	if err != nil {
		return false, err
	}

	caCert, err := x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return false, err
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(caCert)

	keyPair, err = tls.X509KeyPair(certs.ClientCert, certs.ClientKey)
	if err != nil {
		return false, err
	}

	dialer := &net.Dialer{Timeout: 20 * time.Second}
	tlsConfig := &tls.Config{
		RootCAs:            certPool,
		InsecureSkipVerify: false,
		Certificates:       []tls.Certificate{keyPair},
	}

	if _, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConfig); err != nil {
		return false, err
	}

	return true, nil
}

func (stage *Stage) driverOptions() (drivers.DriverOptions, error) {
	options := make(map[string]interface{})

	// Somehow, these are expected by the driver
	options["swarm-master"] = false
	options["swarm-host"] = ""
	options["swarm-discovery"] = ""

	for _, flag := range stage.driver.GetCreateFlags() {
		if flag.Default() == nil {
			options[flag.String()] = false
		} else {
			options[flag.String()] = flag.Default()
		}
	}

	for name, option := range stage.config.Options {
		validOption := false
		driverName := stage.driver.DriverName()
		for _, fuzzyName := range fuzzyOptionNames(driverName, name) {
			if _, ok := options[fuzzyName]; ok {
				options[fuzzyName] = option
				validOption = true
			}
		}
		if !validOption {
			return nil, fmt.Errorf("unsupported option for stage type '%s': '%s'", driverName, name)
		}
	}

	return rpcdriver.RPCFlags{Values: options}, nil
}

func fuzzyOptionNames(driver string, name string) []string {
	prefixed := fmt.Sprintf("%s-%s", driver, name)
	return []string{
		name,
		prefixed,
		strings.ReplaceAll(name, "_", "-"),
		strings.ReplaceAll(prefixed, "_", "-"),
	}
}
