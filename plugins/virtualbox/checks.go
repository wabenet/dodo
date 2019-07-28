package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/oclaussen/go-gimme/ssl"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	retryAttempts  = 60
	retrySleeptime = 4
)

func await(test func() (bool, error)) error {
	for attempts := 0; attempts < retryAttempts; attempts++ {
		success, err := test()
		if err != nil {
			return err
		}
		if success {
			return nil
		}
		time.Sleep(retrySleeptime * time.Second)
	}
	return errors.New("max retries reached")
}

func (vbox *Stage) isVMRunning() (bool, error) {
	stdout, err := vbm("showvminfo", vbox.VMName, "--machinereadable")
	if err != nil {
		return false, err
	}
	re := regexp.MustCompile(`(?m)^VMState="(\w+)"`)
	groups := re.FindStringSubmatch(stdout)
	if len(groups) < 1 {
		return false, errors.New("no vm state in VBoxManage output")
	}
	return groups[1] == "running", nil
}

func (vbox *Stage) isDockerRunning(port int) (bool, error) {
	output, err := vbox.ssh("if ! type netstat 1>/dev/null; then ss -tln; else netstat -tln; fi")
	if err != nil {
		log.Debug("error running SSH command")
	}

	for _, line := range strings.Split(output, "\n") {
		match, err := regexp.MatchString(fmt.Sprintf(":%d\\s+.*:.*", port), line)
		if err != nil {
			return false, err
		}
		if match && line != "" {
			return true, nil
		}
	}

	return false, nil
}

func (vbox *Stage) isPortOpen(port int) (bool, error) {
	ip, err := vbox.GetIP()
	if err != nil {
		return false, err
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), 5*time.Second)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	return true, nil
}

func (vbox *Stage) isDockerResponding(certs *ssl.Certificates) (bool, error) {
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

	dockerURL, err := vbox.GetURL()
	if err != nil {
		return false, err
	}

	parsed, err := url.Parse(dockerURL)
	if err != nil {
		return false, errors.Wrap(err, "could not parse Docker URL")
	}

	if _, err = tls.DialWithDialer(
		&net.Dialer{Timeout: 20 * time.Second},
		"tcp",
		parsed.Host,
		&tls.Config{
			RootCAs:            certPool,
			InsecureSkipVerify: false,
			Certificates:       []tls.Certificate{keyPair},
		},
	); err != nil {
		return false, err
	}

	return true, nil
}
