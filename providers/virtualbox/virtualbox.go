package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/hashicorp/go-plugin"
	"github.com/oclaussen/dodo/pkg/stage/provider"
	"github.com/oclaussen/dodo/providers/virtualbox/boot2docker"
	"github.com/oclaussen/go-gimme/ssh"
	"github.com/oclaussen/go-gimme/ssl"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const defaultPort = 2376 // TODO: get this from docker directly

type Options struct {
	CPU      int
	Memory   int
	DiskSize int
}

type VirtualBoxProvider struct {
	VMName      string
	StoragePath string
	CachePath   string
}

func main() {
	log.SetFormatter(new(log.JSONFormatter))
	plugin.Serve(&plugin.ServeConfig{
		GRPCServer:      plugin.DefaultGRPCServer,
		HandshakeConfig: provider.HandshakeConfig("virtualbox"),
		Plugins: map[string]plugin.Plugin{
			"provider": &provider.ProviderPlugin{
				Impl: &VirtualBoxProvider{},
			},
		},
	})
}

func (vbox *VirtualBoxProvider) Initialize(config map[string]string) (bool, error) {
	vbox.VMName = config["vmName"] // TODO: check if these exist
	vbox.StoragePath = config["storagePath"]
	vbox.CachePath = config["cachePath"]

	if err := checkVBoxManageVersion(); err != nil {
		return false, err
	}

	if _, err := ListHostOnlyAdapters(); err != nil {
		return false, err
	}

	return true, nil
}

func (vbox *VirtualBoxProvider) Create() error {
	opts := Options{CPU: 1, Memory: 1024, DiskSize: 20000}

	if err := boot2docker.UpdateISOCache(vbox.CachePath); err != nil {
		return err
	}

	if err := os.MkdirAll(vbox.StoragePath, 0700); err != nil {
		return err
	}

	log.Info("copying boot2docker.iso...")
	if err := copyFile(
		filepath.Join(vbox.CachePath, "boot2docker.iso"),
		filepath.Join(vbox.StoragePath, "boot2docker.iso"),
	); err != nil {
		return err
	}

	log.Info("creating SSH key...")
	if _, err := ssh.GimmeKeyPair(filepath.Join(vbox.StoragePath, "id_rsa")); err != nil {
		return errors.Wrap(err, "could not generate SSH key")
	}

	log.Info("creating disk image...")
	tarBuf, err := boot2docker.MakeDiskImage(filepath.Join(vbox.StoragePath, "id_rsa.pub"))
	if err != nil {
		return errors.Wrap(err, "could not create disk tarball")
	}

	if err := CreateDiskImage(filepath.Join(vbox.StoragePath, "disk.vmdk"), opts.DiskSize, tarBuf); err != nil {
		return errors.Wrap(err, "cloud not create disk image")
	}

	if _, err := vbm(
		"createvm",
		"--basefolder", vbox.StoragePath,
		"--name", vbox.VMName,
		"--register",
	); err != nil {
		return errors.Wrap(err, "could not create VM")
	}

	cpus := opts.CPU
	if cpus < 1 {
		cpus = int(runtime.NumCPU())
	}
	if cpus > 32 {
		cpus = 32
	}

	if _, err := vbm(
		"modifyvm", vbox.VMName,
		"--firmware", "bios",
		"--bioslogofadein", "off",
		"--bioslogofadeout", "off",
		"--bioslogodisplaytime", "0",
		"--biosbootmenu", "disabled",
		"--ostype", "Linux26_64",
		"--cpus", fmt.Sprintf("%d", cpus),
		"--memory", fmt.Sprintf("%d", opts.Memory),
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
		"--boot1", "dvd",
	); err != nil {
		return errors.Wrap(err, "could not configure general VM settings")
	}

	if _, err := vbm(
		"modifyvm", vbox.VMName,
		"--nic1", "nat",
		"--nictype1", "82540EM",
		"--cableconnected1", "on",
	); err != nil {
		return errors.Wrap(err, "could not create nat controller")
	}

	if _, err := vbm(
		"storagectl", vbox.VMName,
		"--name", "SATA",
		"--add", "sata",
		"--hostiocache", "on",
	); err != nil {
		return errors.Wrap(err, "could not create SATA controller")
	}

	if _, err := vbm(
		"storageattach", vbox.VMName,
		"--storagectl", "SATA",
		"--port", "0",
		"--device", "0",
		"--type", "dvddrive",
		"--medium", filepath.Join(vbox.StoragePath, "boot2docker.iso"),
	); err != nil {
		return errors.Wrap(err, "could not attach boot2docker iso")
	}

	if _, err := vbm(
		"storageattach", vbox.VMName,
		"--storagectl", "SATA",
		"--port", "1",
		"--device", "0",
		"--type", "hdd",
		"--medium", filepath.Join(vbox.StoragePath, "disk.vmdk"),
	); err != nil {
		return errors.Wrap(err, "could not attach main disk")
	}

	if _, err := vbm(
		"guestproperty", "set", vbox.VMName,
		"/VirtualBox/GuestAdd/SharedFolders/MountPrefix", "/",
	); err != nil {
		return errors.Wrap(err, "could not set mount prefxi")
	}
	if _, err := vbm(
		"guestproperty", "set", vbox.VMName,
		"/VirtualBox/GuestAdd/SharedFolders/MountDir", "/",
	); err != nil {
		return errors.Wrap(err, "could not set mount dir")
	}

	shareName, shareDir := getShareDriveAndName()
	if _, err := os.Stat(shareDir); err != nil && !os.IsNotExist(err) {
		return err
	} else if !os.IsNotExist(err) {
		if _, err := vbm(
			"sharedfolder", "add", vbox.VMName,
			"--name", shareName,
			"--hostpath", shareDir,
			"--automount",
		); err != nil {
			return errors.Wrap(err, "could not mount shared folder")
		}

		if _, err := vbm(
			"setextradata", vbox.VMName,
			"VBoxInternal2/SharedFoldersEnableSymlinksCreate/"+shareName, "1",
		); err != nil {
			return errors.Wrap(err, "could not set shared folder extra data")
		}
	}

	if err := vbox.Start(); err != nil {
		return errors.Wrap(err, "could not start VM")
	}

	log.Info("provisioning VM...")

	if err := vbox.setHostname(); err != nil {
		return err
	}

	if err := vbox.makeDockerOptionsDir(); err != nil {
		return err
	}

	ip, err := vbox.GetIP()
	if err != nil {
		return err
	}

	certs, files, err := ssl.GimmeCertificates(&ssl.Options{
		Org:          fmt.Sprintf("dodo.%s", vbox.VMName),
		Hosts:        []string{ip, "localhost"},
		WriteToFiles: &ssl.Files{Directory: vbox.StoragePath},
	})
	if err != nil {
		return err
	}

	if err := vbox.stopDocker(); err != nil {
		return err
	}

	if err := vbox.deleteDockerLink(); err != nil {
		return err
	}

	log.Info("copying certs to the VM...")

	if err := vbox.writeRemoteFile(files.CAFile, path.Join(dockerDir, "ca.pem")); err != nil {
		return err
	}
	if err := vbox.writeRemoteFile(files.ServerCertFile, path.Join(dockerDir, "server.pem")); err != nil {
		return err
	}
	if err := vbox.writeRemoteFile(files.ServerKeyFile, path.Join(dockerDir, "server-key.pem")); err != nil {
		return err
	}

	if err := vbox.writeDockerOptions(defaultPort); err != nil {
		return err
	}

	if err := vbox.startDocker(); err != nil {
		return err
	}

	log.Info("waiting for Docker daemon...")
	if err := await(func() (bool, error) {
		ok, err := vbox.isDockerRunning(defaultPort)
		return ok, err
	}); err != nil {
		return errors.Wrap(err, "the Docker daemon did not start successfully")
	}

	log.Info("checking connection...")
	if ok, err := vbox.isPortOpen(defaultPort); err != nil || !ok {
		return errors.Wrap(err, "could not reach docker port")
	}

	if ok, err := vbox.isDockerResponding(certs); err != nil || !ok {
		return errors.Wrap(err, "docker is not responding")
	}

	log.Info("VM is fully provisioned")
	return nil
}

func (vbox *VirtualBoxProvider) Start() error {
	running, err := vbox.Available()
	if err != nil {
		return err
	}

	if running {
		return errors.New("VM is already running")
	} else {
		log.Info("starting VM...")
	}

	sshForwarding := &PortForwarding{
		Name:      "ssh",
		Interface: 1,
		Protocol:  "tcp",
		GuestPort: 22,
	}

	log.Info("configure network...")
	if err := SetupHostOnlyNetwork(vbox.VMName, "192.168.99.1/24"); err != nil {
		return errors.Wrap(err, "could not set up host-only network")
	}

	if err := ConfigurePortForwarding(vbox.VMName, sshForwarding); err != nil {
		return errors.Wrap(err, "could not configure port forwarding")
	}

	if _, err := vbm("startvm", vbox.VMName, "--type", "headless"); err != nil {
		return errors.Wrap(err, "could not start VM")
	}

	log.Info("waiting for an IP...")
	if err = await(func() (bool, error) {
		ip, err := vbox.GetIP()
		return err == nil && ip != "", nil
	}); err != nil {
		return err
	}

	log.Info("waiting for Docker daemon...")
	if err := await(func() (bool, error) {
		ok, err := vbox.isDockerRunning(defaultPort)
		return ok, err
	}); err != nil {
		return errors.Wrap(err, "the Docker daemon did not start successfully")
	}

	return nil
}

func (vbox *VirtualBoxProvider) Stop() error {
	log.Info("stopping VM...")

	available, err := vbox.Available()
	if err != nil {
		return err
	}
	if !available {
		log.Info("VM is already stopped")
		return nil
	}

	if _, err := vbm("controlvm", vbox.VMName, "acpipowerbutton"); err != nil {
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

func (vbox *VirtualBoxProvider) Remove(force bool) error {
	// TODO: log errors if force=true
	exist, err := vbox.Exist()
	if err != nil && !force {
		return err
	}

	if !exist && !force {
		log.Info("VM does not exist")
		return nil
	}

	log.Info("removing VM...")

	running, err := vbox.Available()
	if err != nil && !force {
		return err
	}

	if running {
		if _, err := vbm("controlvm", vbox.VMName, "poweroff"); err != nil && !force {
			return err
		}
	}

	if _, err = vbm("unregistervm", "--delete", vbox.VMName); err != nil && !force {
		return err
	}

	if err := os.RemoveAll(vbox.StoragePath); err != nil && !force {
		return errors.Wrap(err, "could not remove storage dir")
	}

	log.Info("removed VM")
	return nil
}

func (vbox *VirtualBoxProvider) Exist() (bool, error) {
	if _, err := os.Stat(vbox.StoragePath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	_, err := vbm("showvminfo", vbox.VMName, "--machinereadable")
	return err == nil, nil
}

func (vbox *VirtualBoxProvider) Available() (bool, error) {
	return vbox.isVMRunning()
}

func (vbox *VirtualBoxProvider) GetURL() (string, error) {
	ip, err := vbox.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", errors.New("could not get IP")
	}
	return fmt.Sprintf("tcp://%s:%d", ip, defaultPort), nil
}

func (vbox *VirtualBoxProvider) GetIP() (string, error) {
	running, err := vbox.Available()
	if err != nil {
		return "", err
	}
	if !running {
		return "", errors.New("VM is not running")
	}

	stdout, err := vbm("showvminfo", vbox.VMName, "--machinereadable")
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`(?m)^hostonlyadapter([\d]+)`)
	groups := re.FindStringSubmatch(stdout)
	if len(groups) < 2 {
		return "", errors.New("VM does not have a host-only adapter")
	}

	re = regexp.MustCompile(fmt.Sprintf("(?m)^macaddress%s=\"(.*)\"", groups[1]))
	groups = re.FindStringSubmatch(stdout)
	if len(groups) < 2 {
		return "", errors.New("could not find MAC address for host-only adapter")
	}

	macAddress := strings.ToLower(groups[1])

	output, err := vbox.ssh("ip addr show")
	if err != nil {
		return "", err
	}

	inTargetMacBlock := false
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "link") {
			values := strings.Split(line, " ")
			if len(values) >= 2 {
				if strings.Replace(values[1], ":", "", -1) == macAddress {
					inTargetMacBlock = true
				}
			}

		} else if inTargetMacBlock && strings.HasPrefix(line, "inet") && !strings.HasPrefix(line, "inet6") {
			values := strings.Split(line, " ")
			if len(values) >= 2 {
				return values[1][:strings.Index(values[1], "/")], nil
			}
		}
	}

	return "", errors.New("could not find IP")
}

func (vbox *VirtualBoxProvider) GetDockerOptions() (*provider.DockerOptions, error) {
	url, err := vbox.GetURL()
	if err != nil {
		return nil, err
	}
	return &provider.DockerOptions{
		Host:     url,
		CAFile:   filepath.Join(vbox.StoragePath, "ca.pem"),
		CertFile: filepath.Join(vbox.StoragePath, "client.pem"),
		KeyFile:  filepath.Join(vbox.StoragePath, "client-key.pem"),
	}, nil
}
