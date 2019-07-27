package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/oclaussen/dodo/pkg/stage/boot2docker"
	"github.com/oclaussen/dodo/pkg/stage/provider"
	"github.com/oclaussen/go-gimme/ssh"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type VirtualBoxProvider struct {
	VMName      string
	StoragePath string
}

func main() {
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

	if err := checkVBoxManageVersion(); err != nil {
		return false, err
	}

	if _, err := ListHostOnlyAdapters(); err != nil {
		return false, err
	}

	return true, nil
}

func (vbox *VirtualBoxProvider) Create() error {
	opts := provider.Options{CPU: 1, Memory: 1024, DiskSize: 20000}
	log.Info("creating VirtualBox VM...")

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

	return nil
}

func (vbox *VirtualBoxProvider) Start() error {
	status, err := vbox.Status()
	if err != nil {
		return err
	}

	sshForwarding := &PortForwarding{
		Name:      "ssh",
		Interface: 1,
		Protocol:  "tcp",
		GuestPort: 22,
	}

	if status != provider.Paused {
		return errors.New("VM not in startable state")
	}

	log.Info("check network to re-create if needed...")
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
	for i := 0; i < 60; i++ {
		if ip, err := vbox.GetIP(); err == nil && ip != "" {
			return nil
		}
		time.Sleep(4 * time.Second)
	}

	return errors.New("could not get IP address")

}

func (vbox *VirtualBoxProvider) Stop() error {
	status, err := vbox.Status()
	if err != nil {
		return err
	}

	if _, err := vbm("controlvm", vbox.VMName, "acpipowerbutton"); err != nil {
		return err
	}
	for {
		status, err = vbox.Status()
		if err != nil {
			return err
		}
		if status == provider.Up {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	return nil
}

func (vbox *VirtualBoxProvider) Remove() error {
	status, err := vbox.Status()
	if err != nil {
		return err
	}

	if status != provider.Paused {
		if _, err := vbm("controlvm", vbox.VMName, "poweroff"); err != nil {
			return err
		}
	}

	_, err = vbm("unregistervm", "--delete", vbox.VMName)
	return err
}

func (vbox *VirtualBoxProvider) GetURL() (string, error) {
	ip, err := vbox.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", errors.New("could not get IP")
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (vbox *VirtualBoxProvider) GetIP() (string, error) {
	status, err := vbox.Status()
	if err != nil {
		return "", err
	}
	if status != provider.Up {
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

	opts, err := vbox.GetSSHOptions()
	if err != nil {
		return "", err
	}

	executor, err := ssh.GimmeExecutor(&ssh.Options{
		Host:              opts.Hostname,
		Port:              opts.Port,
		User:              opts.Username,
		IdentityFileGlobs: []string{filepath.Join(vbox.StoragePath, "id_rsa")},
		NonInteractive:    true,
	})
	if err != nil {
		return "", nil
	}
	defer executor.Close()
	output, err := executor.Execute("ip addr show")

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
