package stage

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
)

const defaultPort = 2376

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

func (stage *Stage) writeRemoteFile(localPath string, remotePath string) error {
	bytes, err := ioutil.ReadFile(localPath)
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf("printf '%%s' '%s' | sudo tee %s", string(bytes), remotePath)
	if _, err := stage.RunSSHCommand(cmd); err != nil {
		return err
	}

	return nil
}

func (stage *Stage) setHostname() error {
	cmd := fmt.Sprintf(
		"sudo /usr/bin/sethostname %s && echo %q | sudo tee /var/lib/boot2docker/etc/hostname",
		stage.name,
		stage.name,
	)
	if _, err := stage.RunSSHCommand(cmd); err != nil {
		return err
	}
	return nil
}

func (stage *Stage) makeDockerOptionsDir() error {
	dockerDir := "/var/lib/boot2docker"
	cmd := fmt.Sprintf("sudo mkdir -p %s", dockerDir)
	if _, err := stage.RunSSHCommand(cmd); err != nil {
		return err
	}
	return nil
}

func (stage *Stage) deleteDockerLink() error {
	cmd := `if [ ! -z "$(ip link show docker0)" ]; then sudo ip link delete docker0; fi`
	if _, err := stage.RunSSHCommand(cmd); err != nil {
		return err
	}
	return nil
}

func (stage *Stage) startDocker() error {
	cmd := "sudo /etc/init.d/docker start"
	if _, err := stage.RunSSHCommand(cmd); err != nil {
		return err
	}
	return nil
}

func (stage *Stage) stopDocker() error {
	cmd := "sudo /etc/init.d/docker stop"
	if _, err := stage.RunSSHCommand(cmd); err != nil {
		return err
	}
	return nil
}

func (stage *Stage) writeDockerOptions(dockerPort int) error {
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

	dockerDir := "/var/lib/boot2docker"
	targetPath := path.Join(dockerDir, "profile")
	cmd := fmt.Sprintf("printf '%%s' \"%s\" | sudo tee %s", config.String(), targetPath)
	if _, err := stage.RunSSHCommand(cmd); err != nil {
		return err
	}

	return nil
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
