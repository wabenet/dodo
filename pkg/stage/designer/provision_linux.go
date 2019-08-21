// +build linux

package designer

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"text/template"
	"time"
)

const (
	dockerPort = 2376
	dataDir    = "/var/lib/boot2docker"

	dockerOptionsTemplate = `
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
)

type dockerOptionsContext struct {
	DockerPort    int
	StorageDriver string
	CACert        string
	ServerCert    string
	ServerKey     string
	Environment   []string
	DockerArgs    []string
}

func Provision(config *Config) error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	if len(config.Hostname) > 0 {
		syscall.Sethostname([]byte(config.Hostname))
		if err := writeDataFile(filepath.Join("etc", "hostname"), config.Hostname); err != nil {
			return err
		}
	}

	if err := exec.Command("/etc/init.d/docker", "stop").Run(); err != nil {
		return err
	}

	if len(config.CA) > 0 {
		if err := writeDataFile("ca.pem", config.CA); err != nil {
			return err
		}
	}
	if len(config.ServerCert) > 0 {
		if err := writeDataFile("server.pem", config.ServerCert); err != nil {
			return err
		}
	}
	if len(config.ServerKey) > 0 {
		if err := writeDataFile("server-key.pem", config.ServerKey); err != nil {
			return err
		}
	}

	tmpl, err := template.New("dockerOptions").Parse(dockerOptionsTemplate)
	if err != nil {
		return err
	}

	var dockerOptions bytes.Buffer
	tmpl.Execute(&dockerOptions, dockerOptionsContext{
		DockerPort:    dockerPort,
		StorageDriver: "overlay2",
		CACert:        filepath.Join(dataDir, "ca.pem"),
		ServerCert:    filepath.Join(dataDir, "server.pem"),
		ServerKey:     filepath.Join(dataDir, "server-key.pem"),
		Environment:   config.Environment,
		DockerArgs:    config.DockerArgs,
	})
	if err := writeDataFile("profile", dockerOptions.String()); err != nil {
		return err
	}

	if err := exec.Command("/etc/init.d/docker", "start").Run(); err != nil {
		return err
	}

	for attempts := 0; attempts < 60; attempts++ {
		if conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", dockerPort)); err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(5 * time.Second)
	}

	return errors.New("docker did not start successfully")
}

func writeDataFile(path string, content string) error {
	return ioutil.WriteFile(filepath.Join(dataDir, path), []byte(content), 0644)
}
