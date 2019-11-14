package designer

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	log "github.com/sirupsen/logrus"
)

const (
	dockerPort      = 2376
	dockerConfigDir = "/etc/docker"

	systemdUnitPath     = "/etc/systemd/system/docker.service"
	systemdUnitTemplate = `[Service]
ExecStart=
ExecStart={{ .DockerdBinary }} -H tcp://0.0.0.0:{{ .DockerPort }} -H unix:///var/run/docker.sock --storage-driver {{ .StorageDriver }} --tlsverify --tlscacert {{ .CACert }} --tlscert {{ .ServerCert }} --tlskey {{ .ServerKey }} {{ range .DockerArgs }}--{{.}} {{ end }}
Environment={{range .Environment }}{{ printf "%q" . }} {{end}}
`

	genericOptionsPath     = "/etc/docker/profile"
	genericOptionsTemplate = `
DOCKER_OPTS='-H tcp://0.0.0.0:{{ .DockerPort }} -H unix:///var/run/docker.sock --storage-driver {{ .StorageDriver }} --tlsverify --tlscacert {{ .CACert }} --tlscert {{ .ServerCert }} --tlskey {{ .ServerKey }}{{ range .DockerArgs}}--{{.}} {{ end }}'
{{range .Environment }}export \"{{ printf "%q" . }}\"
{{end}}
`
)

type DockerConfig struct {
	CA          []byte
	ServerCert  []byte
	ServerKey   []byte
	Environment []string
	Arguments   []string
}

type dockerOptionsContext struct {
	DockerdBinary string
	DockerPort    int
	StorageDriver string
	CACert        string
	ServerCert    string
	ServerKey     string
	Environment   []string
	DockerArgs    []string
}

func InstallDocker() error {
	if pacman, err := exec.LookPath("pacman"); err == nil {
		return exec.Command(pacman, "-Sy", "--noconfirm", "--noprogressbar", "docker").Run()
	} else if zypper, err := exec.LookPath("zypper"); err == nil {
		return exec.Command(zypper, "-n", "in", "docker").Run()
	} else if yum, err := exec.LookPath("yum"); err == nil {
		return exec.Command(yum, "install", "-y", "docker").Run()
	} else if aptget, err := exec.LookPath("apt-get"); err == nil {
		if err := exec.Command(aptget, "update").Run(); err != nil {
			return err
		}
		aptcache, err := exec.LookPath("apt-cache")
		if err != nil {
			return err
		}
		for _, pkg := range []string{"docker-ce", "docker.io", "docker-engine", "docker"} {
			out, err := exec.Command(aptcache, "show", "-q", pkg).Output()
			if err == nil && len(out) > 0 {
				cmd := exec.Command(aptget, "install", "-y", pkg)
				cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
				return cmd.Run()
			}
		}
	}
	log.Warn("no valid docker installation method found, assuming it is already installed")
	return nil
}

func RestartDocker() error {
	if systemctl, err := exec.LookPath("systemctl"); err == nil {
		if err := exec.Command(systemctl, "daemon-reload").Run(); err != nil {
			return err
		}
		if err := exec.Command(systemctl, "-f", "restart", "docker").Run(); err != nil {
			return err
		}
		if err := exec.Command(systemctl, "-f", "enable", "docker").Run(); err != nil {
			return err
		}
		return nil
	} else if service, err := exec.LookPath("service"); err == nil {
		return exec.Command(service, "docker", "restart").Run()
	}
	log.Warn("could not start docker daemon")
	return nil
}

func ConfigureDocker(config *DockerConfig) error {
	caPath := filepath.Join(dockerConfigDir, "ca.pem")
	certPath := filepath.Join(dockerConfigDir, "server.pem")
	keyPath := filepath.Join(dockerConfigDir, "server-key.pem")

	if err := os.MkdirAll(dockerConfigDir, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(caPath, config.CA, 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile(certPath, config.ServerCert, 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile(keyPath, config.ServerKey, 0644); err != nil {
		return err
	}

	dockerd, err := exec.LookPath("dockerd")
	if err != nil {
		return err
	}

	context := dockerOptionsContext{
		DockerdBinary: dockerd,
		DockerPort:    dockerPort,
		StorageDriver: "overlay2",
		CACert:        caPath,
		ServerCert:    certPath,
		ServerKey:     keyPath,
		Environment:   config.Environment,
		DockerArgs:    config.Arguments,
	}

	if _, err := exec.LookPath("systemctl"); err == nil {
		tmpl, err := template.New("systemd").Parse(systemdUnitTemplate)
		if err != nil {
			return err
		}

		var content bytes.Buffer
		tmpl.Execute(&content, context)
		if err := ioutil.WriteFile(systemdUnitPath, content.Bytes(), 0644); err != nil {
			return err
		}
	} else {
		tmpl, err := template.New("dockerOptions").Parse(genericOptionsTemplate)
		if err != nil {
			return err
		}

		var content bytes.Buffer
		tmpl.Execute(&content, context)
		if err := ioutil.WriteFile(genericOptionsPath, content.Bytes(), 0644); err != nil {
			return err
		}
	}
	return nil
}
