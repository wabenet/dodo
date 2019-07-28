package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"text/template"

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

func (vbox *Stage) writeRemoteFile(localPath string, remotePath string) error {
	bytes, err := ioutil.ReadFile(localPath)
	if err != nil {
		return err
	}

	_, err = vbox.ssh(fmt.Sprintf("printf '%%s' '%s' | sudo tee %s", string(bytes), remotePath))
	return err
}

func (vbox *Stage) setHostname() error {
	_, err := vbox.ssh(fmt.Sprintf(
		"sudo /usr/bin/sethostname %s && echo %q | sudo tee /var/lib/boot2docker/etc/hostname",
		vbox.VMName,
		vbox.VMName,
	))
	return err
}

func (vbox *Stage) makeDockerOptionsDir() error {
	_, err := vbox.ssh(fmt.Sprintf("sudo mkdir -p %s", dockerDir))
	return err
}

func (vbox *Stage) deleteDockerLink() error {
	_, err := vbox.ssh(`if [ ! -z "$(ip link show docker0)" ]; then sudo ip link delete docker0; fi`)
	return err
}

func (vbox *Stage) startDocker() error {
	_, err := vbox.ssh("sudo /etc/init.d/docker start")
	return err
}

func (vbox *Stage) stopDocker() error {
	_, err := vbox.ssh("sudo /etc/init.d/docker stop")
	return err
}

func (vbox *Stage) writeDockerOptions(dockerPort int) error {
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
