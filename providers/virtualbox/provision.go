package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/oclaussen/go-gimme/ssl"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const dockerDir = "/var/lib/boot2docker"

const dockerOptionsTemplate = `
EXTRA_ARGS='
{{ range .DockerArgs }}--{{ . }}
{{ end }}
'
DOCKER_HOST='-H tcp://0.0.0.0:{{ .DockerPort }}'
DOCKER_STORAGE={{ .StorageDriver }}
DOCKER_TLS=auto
CACERT={{ .CACert }}
SERVERCERT={{ .ServerCert }}
SERVERKEY={{ .ServerKey }}

{{range .Environment}}export \"{{ printf "%q" . }}\"
{{end}}
`

const genericOptionsTemplate = `
DOCKER_OPTS='
-H tcp://0.0.0.0:{{ .DockerPort }}
-H unix:///var/run/docker.sock
--storage-driver {{ .StorageDriver }}
--tlsverify
--tlscacert {{ .CACert }}
--tlscert {{ .ServerCert }}
--tlskey {{ .ServerKey }}
{{ range .DockerArgs }}--{{ . }}
{{ end }}
'
{{ range .Environment }}export \"{{ printf "%q" . }}\"
{{ end }}
`

type dockerOptionsContext struct {
	DockerPort    int
	StorageDriver string
	CACert        string
	ServerCert    string
	ServerKey     string
	Environment   []string
	DockerArgs    []string
}

func (vbox *VirtualBoxProvider) writeRemoteFile(localPath string, remotePath string) error {
	bytes, err := ioutil.ReadFile(localPath)
	if err != nil {
		return err
	}

	_, err = vbox.ssh(fmt.Sprintf("printf '%%s' '%s' | sudo tee %s", string(bytes), remotePath))
	return err
}

func (vbox *VirtualBoxProvider) setHostname() error {
	_, err := vbox.ssh(fmt.Sprintf(
		"sudo /usr/bin/sethostname %s && echo %q | sudo tee /var/lib/boot2docker/etc/hostname",
		vbox.VMName,
		vbox.VMName,
	))
	return err
}

func (vbox *VirtualBoxProvider) makeDockerOptionsDir() error {
	_, err := vbox.ssh(fmt.Sprintf("sudo mkdir -p %s", dockerDir))
	return err
}

func (vbox *VirtualBoxProvider) deleteDockerLink() error {
	_, err := vbox.ssh(`if [ ! -z "$(ip link show docker0)" ]; then sudo ip link delete docker0; fi`)
	return err
}

func (vbox *VirtualBoxProvider) startDocker() error {
	_, err := vbox.ssh("sudo /etc/init.d/docker start")
	return err
}

func (vbox *VirtualBoxProvider) stopDocker() error {
	_, err := vbox.ssh("sudo /etc/init.d/docker stop")
	return err
}

func (vbox *VirtualBoxProvider) writeDockerOptions(dockerPort int) error {
	tmpl, err := template.New("engineConfig").Parse(dockerOptionsTemplate)
	if err != nil {
		return err
	}

	var config bytes.Buffer
	// TODO: these should come from configuration
	tmpl.Execute(&config, dockerOptionsContext{
		DockerPort:    dockerPort,
		StorageDriver: "overlay2",
		CACert:        path.Join(dockerDir, "ca.pem"),
		ServerCert:    path.Join(dockerDir, "server.pem"),
		ServerKey:     path.Join(dockerDir, "server-key.pem"),
		Environment:   []string{},
		DockerArgs:    []string{},
	})

	log.Info("writing Docker configuration on the remote daemon...")

	targetPath := path.Join(dockerDir, "profile")
	_, err = vbox.ssh(fmt.Sprintf("printf '%%s' \"%s\" | sudo tee %s", config.String(), targetPath))
	return err
}

func parseDockerPort(dockerURL string) (int, error) {
	parsed, err := url.Parse(dockerURL)
	if err != nil {
		return 0, err
	}

	parts := strings.Split(parsed.Host, ":")
	if len(parts) == 2 {
		return strconv.Atoi(parts[1])
	}

	return defaultPort, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}

	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	fi, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, fi.Mode())
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

func (vbox *VirtualBoxProvider) waitForDocker() error {
	maxAttempts := 60
	reDaemonListening := fmt.Sprintf(":%d\\s+.*:.*", defaultPort)
	cmd := "if ! type netstat 1>/dev/null; then ss -tln; else netstat -tln; fi"

	log.Info("waiting for Docker daemon...")
	for i := 0; i < maxAttempts; i++ {
		output, err := vbox.ssh(cmd)
		if err != nil {
			log.WithFields(log.Fields{"cmd": cmd}).Debug("error running SSH command")
		}

		for _, line := range strings.Split(output, "\n") {
			match, err := regexp.MatchString(reDaemonListening, line)
			if err != nil {
				log.Warnf("Regex warning: %s", err)
			}
			if match && line != "" {
				return nil
			}
		}

		time.Sleep(3 * time.Second)

	}
	return errors.New("the Docker daemon did not start successfully")
}
