package stage

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/oclaussen/go-gimme/ssl"
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

func (stage *Stage) Provision() error {
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

	_, files, err := ssl.GimmeCertificates(&ssl.Options{
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

	return nil
}

func (stage *Stage) writeRemoteFile(localPath string, remotePath string) error {
	bytes, err := ioutil.ReadFile(localPath)
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf("printf '%%s' '%s' | sudo tee %s", string(bytes), remotePath)
	if _, err := drivers.RunSSHCommandFromDriver(stage.driver, cmd); err != nil {
		return err
	}

	return nil
}

func (stage *Stage) setHostname() error {
	hostname := stage.driver.GetMachineName()
	cmd := fmt.Sprintf(
		"sudo /usr/bin/sethostname %s && echo %q | sudo tee /var/lib/boot2docker/etc/hostname",
		hostname,
		hostname,
	)
	if _, err := drivers.RunSSHCommandFromDriver(stage.driver, cmd); err != nil {
		return err
	}
	return nil
}

func (stage *Stage) makeDockerOptionsDir() error {
	dockerDir := "/var/lib/boot2docker"
	cmd := fmt.Sprintf("sudo mkdir -p %s", dockerDir)
	if _, err := drivers.RunSSHCommandFromDriver(stage.driver, cmd); err != nil {
		return err
	}
	return nil
}

func (stage *Stage) deleteDockerLink() error {
	cmd := `if [ ! -z "$(ip link show docker0)" ]; then sudo ip link delete docker0; fi`
	if _, err := drivers.RunSSHCommandFromDriver(stage.driver, cmd); err != nil {
		return err
	}
	return nil
}

func (stage *Stage) startDocker() error {
	cmd := "sudo /etc/init.d/docker start"
	if _, err := drivers.RunSSHCommandFromDriver(stage.driver, cmd); err != nil {
		return err
	}
	return nil
}

func (stage *Stage) stopDocker() error {
	cmd := "sudo /etc/init.d/docker stop"
	if _, err := drivers.RunSSHCommandFromDriver(stage.driver, cmd); err != nil {
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
		DockerArgs: []string{
			fmt.Sprintf("--label provider=%s", stage.driver.DriverName()),
		},
	})

	log.Info("writing Docker configuration on the remote daemon...")

	dockerDir := "/var/lib/boot2docker"
	targetPath := path.Join(dockerDir, "profile")
	cmd := fmt.Sprintf("printf '%%s' \"%s\" | sudo tee %s", config.String(), targetPath)
	if _, err := drivers.RunSSHCommandFromDriver(stage.driver, cmd); err != nil {
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

	return engine.DefaultPort, nil
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
