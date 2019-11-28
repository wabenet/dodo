package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"github.com/oclaussen/dodo/pkg/integrations/ova"
	"github.com/oclaussen/dodo/pkg/integrations/virtualbox"
	"github.com/oclaussen/dodo/pkg/stage"
	"github.com/oclaussen/dodo/pkg/stage/box"
	"github.com/oclaussen/dodo/pkg/stage/designer"
	"github.com/oclaussen/dodo/pkg/stage/provision"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/oclaussen/go-gimme/ssh"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const defaultPort = 2376 // TODO: get this from docker directly

type Stage struct {
	VM          *virtualbox.VM
	Box         *types.Box
	State       *State
	StoragePath string
}

type Options struct {
	CPU    int
	Memory int
}

func (vbox *Stage) Initialize(name string, config *types.Stage) (bool, error) {
	if vmName, ok := config.Options["name"]; ok {
		vbox.VM = &virtualbox.VM{Name: vmName}
	} else {
		vbox.VM = &virtualbox.VM{Name: name}
	}

	vbox.Box = &config.Box

	baseDir := filepath.FromSlash("/var/lib/dodo") // TODO: choose a better default
	if path, ok := config.Options["path"]; ok {
		baseDir = path
	} else if user, err := user.Current(); err == nil && user.HomeDir != "" {
		baseDir = filepath.Join(user.HomeDir, ".dodo")
	}

	vbox.StoragePath = filepath.Join(baseDir, "stages", name)

	if err := vbox.loadState(); err != nil {
		return false, err
	}

	return true, nil
}

func (vbox *Stage) Create() error {
	opts := Options{CPU: 1, Memory: 1024}

	if err := os.MkdirAll(vbox.StoragePath, 0700); err != nil {
		return err
	}

	log.Info("creating SSH key...")
	if _, err := ssh.GimmeKeyPair(filepath.Join(vbox.StoragePath, "id_rsa")); err != nil {
		return errors.Wrap(err, "could not generate SSH key")
	}

	b, err := box.Load(vbox.Box, "virtualbox")
	if err != nil {
		return errors.Wrap(err, "could not load box")
	}
	if err := b.Download(); err != nil {
		return errors.Wrap(err, "could not download box")
	}

	sshOpts, err := b.GetSSHOptions()
	if err != nil {
		return err
	}

	vbox.State = &State{
		Username:       sshOpts.Username,
		PrivateKeyFile: sshOpts.PrivateKeyFile,
	}
	if err := vbox.saveState(); err != nil {
		return err
	}

	boxFile := filepath.Join(b.Path(), "box.ovf")
	ovf, err := ova.ReadOVF(boxFile)
	if err != nil {
		return err
	}
	importArgs := []string{boxFile, "--vsys", "0", "--vmname", vbox.VM.Name, "--basefolder", vbox.StoragePath}

	for _, item := range ovf.VirtualSystem.VirtualHardware.Items {
		switch item.ResourceType {
		case ova.TypeCPU:
			var cpus string
			if opts.CPU < 1 {
				cpus = fmt.Sprintf("%d", runtime.NumCPU())
			} else if opts.CPU > 32 {
				cpus = "32"
			} else {
				cpus = fmt.Sprintf("%d", opts.CPU)
			}
			importArgs = append(importArgs, "--vsys", "0", "--cpus", cpus)
		case ova.TypeMemory:
			importArgs = append(importArgs, "--vsys", "0", "--memory", fmt.Sprintf("%d", opts.Memory))
		}
	}

	if err := vbox.VM.Import(importArgs...); err != nil {
		return errors.Wrap(err, "could not import VM")
	}

	if err := vbox.VM.Modify(
		"--firmware", "bios",
		"--bioslogofadein", "off",
		"--bioslogofadeout", "off",
		"--bioslogodisplaytime", "0",
		"--biosbootmenu", "disabled",
		"--acpi", "on",
		"--ioapic", "on",
		"--rtcuseutc", "on",
		"--natdnshostresolver1", "off",
		"--natdnsproxy1", "on",
		"--cpuhotplug", "off",
		"--pae", "on",
		"--hpet", "on",
		"--hwvirtex", "on",
		"--nestedpaging", "on",
		"--largepages", "on",
		"--vtxvpid", "on",
		"--accelerate3d", "off",
	); err != nil {
		return errors.Wrap(err, "could not configure general VM settings")
	}

	if err := vbox.VM.Modify(
		"--nic1", "nat",
		"--nictype1", "82540EM",
		"--cableconnected1", "on",
	); err != nil {
		return errors.Wrap(err, "could not create nat controller")
	}

	if err := vbox.VM.ConfigureSharedFolders(getSharedFolders()); err != nil {
		return err
	}

	return vbox.Start()
}

func (vbox *Stage) Start() error {
	running, err := vbox.Available()
	if err != nil {
		return err
	}

	if running {
		return errors.New("VM is already running")
	}
	log.Info("starting VM...")

	log.Info("configure network...")
	if err := vbox.SetupHostOnlyNetwork("192.168.99.1/24"); err != nil {
		return errors.Wrap(err, "could not set up host-only network")
	}

	sshForwarding := vbox.VM.NewPortForwarding("ssh")
	sshForwarding.GuestPort = 22
	if err := sshForwarding.Create(); err != nil {
		return errors.Wrap(err, "could not configure port forwarding")
	}

	if err := vbox.VM.Start(); err != nil {
		return errors.Wrap(err, "could not start VM")
	}

	log.Info("waiting for SSH...")
	if err = await(vbox.isSSHAvailable); err != nil {
		return err
	}

	sshOpts, err := vbox.GetSSHOptions()
	if err != nil {
		return err
	}

	// TODO: replace ssh key
	publicKey, err := ioutil.ReadFile(filepath.Join(vbox.StoragePath, "id_rsa.pub"))
	if err != nil {
		return err
	}

	provisionConfig := &designer.Config{
		Hostname:          vbox.VM.Name,
		DefaultUser:       sshOpts.Username,
		AuthorizedSSHKeys: []string{string(publicKey)},
	}

	result, err := provision.Provision(sshOpts, provisionConfig)
	if err != nil {
		return err
	}

	vbox.State.IPAddress = result.IPAddress
	vbox.State.PrivateKeyFile = filepath.Join(vbox.StoragePath, "id_rsa")
	if err := vbox.saveState(); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(vbox.StoragePath, "ca.pem"), []byte(result.CA), 0600); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(vbox.StoragePath, "client.pem"), []byte(result.ClientCert), 0600); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(vbox.StoragePath, "client-key.pem"), []byte(result.ClientKey), 0600); err != nil {
		return err
	}

	pemData, _ := pem.Decode([]byte(result.CA))
	caCert, err := x509.ParseCertificate(pemData.Bytes)
	if err != nil {
		return err
	}
	certPool := x509.NewCertPool()
	certPool.AddCert(caCert)

	keyPair, err := tls.X509KeyPair([]byte(result.ClientCert), []byte(result.ClientKey))
	if err != nil {
		return err
	}

	dockerURL, err := vbox.GetURL()
	if err != nil {
		return err
	}
	parsed, err := url.Parse(dockerURL)
	if err != nil {
		return errors.Wrap(err, "could not parse Docker URL")
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
		return err
	}

	log.Info("VM is fully provisioned and running")
	return nil
}

func (vbox *Stage) Stop() error {
	log.Info("stopping VM...")

	available, err := vbox.Available()
	if err != nil {
		return err
	}
	if !available {
		log.Info("VM is already stopped")
		return nil
	}

	if err := vbox.VM.Stop(false); err != nil {
		return err
	}

	if err := await(func() (bool, error) {
		available, err := vbox.Available()
		return !available, err
	}); err != nil {
		return err
	}

	return errors.New("VM did not stop successfully")
}

func (vbox *Stage) Remove(force bool) error {
	exist, err := vbox.Exist()
	if err != nil {
		if force {
			log.Error(err)
		} else {
			return err
		}
	}

	if !exist && !force {
		log.Info("VM does not exist")
		return nil
	}

	log.Info("removing VM...")

	running, err := vbox.Available()
	if err != nil {
		if force {
			log.Error(err)
		} else {
			return err
		}
	}

	if running {
		if err := vbox.VM.Stop(true); err != nil {
			if force {
				log.Error(err)
			} else {
				return err
			}
		}
	}

	if err = vbox.VM.Delete(); err != nil {
		if force {
			log.Error(err)
		} else {
			return err
		}
	}

	if err := os.RemoveAll(vbox.StoragePath); err != nil {
		if force {
			log.Error(err)
		} else {
			return err
		}
	}

	log.Info("removed VM")
	return nil
}

func (vbox *Stage) Exist() (bool, error) {
	if _, err := os.Stat(vbox.StoragePath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	_, err := vbox.VM.Info()
	return err == nil, nil
}

func (vbox *Stage) Available() (bool, error) {
	info, err := vbox.VM.Info()
	if err != nil {
		return false, err
	}
	state, ok := info["VMState"]
	return ok && state == "running", nil
}

func (vbox *Stage) GetURL() (string, error) {
	return fmt.Sprintf("tcp://%s:%d", vbox.State.IPAddress, defaultPort), nil
}

func (vbox *Stage) GetSSHOptions() (*stage.SSHOptions, error) {
	portForwardings, err := vbox.VM.ListPortForwardings()
	if err != nil {
		return nil, err
	}

	port := 0
	for _, forward := range portForwardings {
		if forward.Name == "ssh" {
			port = forward.HostPort
			break
		}
	}
	if port == 0 {
		return nil, errors.New("no port forwarding matching ssh port found")
	}

	return &stage.SSHOptions{
		Hostname:       "127.0.0.1",
		Port:           port,
		Username:       vbox.State.Username,
		PrivateKeyFile: vbox.State.PrivateKeyFile,
	}, nil
}

func (vbox *Stage) GetDockerOptions() (*stage.DockerOptions, error) {
	url, err := vbox.GetURL()
	if err != nil {
		return nil, err
	}
	return &stage.DockerOptions{
		Host:     url,
		CAFile:   filepath.Join(vbox.StoragePath, "ca.pem"),
		CertFile: filepath.Join(vbox.StoragePath, "client.pem"),
		KeyFile:  filepath.Join(vbox.StoragePath, "client-key.pem"),
	}, nil
}
