package stage

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/oclaussen/dodo/pkg/stage/boot2docker"
	vbox "github.com/oclaussen/dodo/pkg/stage/virtualbox"
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

	currentStatus, err := vbox.GetStatus(stage.name)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Debug("could not get machine status")
	}
	if currentStatus == vbox.Running {
		log.WithFields(log.Fields{"name": stage.name}).Info("stage is already up")
		return nil
	}

	if err := vbox.Start(stage.name, filepath.Join(stage.stateDir, "machines", stage.name)); err != nil {
		return errors.Wrap(err, "could not start stage")
	}

	if err := stage.waitForDocker(); err != nil {
		return errors.Wrap(err, "docker did start successfully")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("stage is now up")
	return nil

}

func (stage *Stage) create() error {
	log.WithFields(log.Fields{"name": stage.name}).Info("running pre-create checks...")
	if err := vbox.PreCreateCheck(); err != nil {
		return err
	}

	if err := boot2docker.UpdateISOCache(filepath.Join(stage.stateDir, "cache")); err != nil {
		return err
	}

	if err := os.MkdirAll(stage.hostDir(), 0700); err != nil {
		return err
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("copying boot2docker.iso...")
	if err := copyFile(
		filepath.Join(stage.stateDir, "cache", "boot2docker.iso"),
		filepath.Join(stage.hostDir(), "boot2docker.iso"),
	); err != nil {
		return err
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("creating stage...")
	opts := vbox.Options{CPU: 1, Memory: 1024, DiskSize: 20000}
	if err := vbox.Create(stage.name, filepath.Join(stage.stateDir, "machines", stage.name), opts); err != nil {
		return errors.Wrap(err, "could not create stage")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("starting stage...")
	if err := vbox.Start(stage.name, filepath.Join(stage.stateDir, "machines", stage.name)); err != nil {
		return errors.Wrap(err, "could not start stage")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("waiting for stage...")
	if err := stage.waitForStatus(vbox.Running); err != nil {
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

	ip, err := vbox.GetIP(stage.name, filepath.Join(stage.stateDir, "machines", stage.name))
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

	dockerURL, err := vbox.GetURL(stage.name, filepath.Join(stage.stateDir, "machines", stage.name))
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
