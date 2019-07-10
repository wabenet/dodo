package virtualbox

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/oclaussen/dodo/pkg/stage/boot2docker"
	"github.com/oclaussen/go-gimme/ssh"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Options struct {
	CPU      int
	Memory   int
	DiskSize int
}

func PreCreateCheck() error {
	if err := checkVBoxManageVersion(); err != nil {
		return err
	}

	if _, err := ListHostOnlyAdapters(); err != nil {
		return err
	}

	return nil
}

func Create(name string, path string, opts Options) error {
	log.Info("creating VirtualBox VM...")

	log.Info("creating SSH key...")
	if _, err := ssh.GimmeKeyPair(filepath.Join(path, "id_rsa")); err != nil {
		return err
	}

	log.Info("creating disk image...")
	tarBuf, err := boot2docker.MakeDiskImage(filepath.Join(path, "id_rsa.pub"))
	if err != nil {
		return err
	}

	if err := CreateDiskImage(filepath.Join(path, "disk.vmdk"), opts.DiskSize, tarBuf); err != nil {
		return err
	}

	if _, err := vbm(
		"createvm",
		"--basefolder", path,
		"--name", name,
		"--register",
	); err != nil {
		return err
	}

	cpus := opts.CPU
	if cpus < 1 {
		cpus = int(runtime.NumCPU())
	}
	if cpus > 32 {
		cpus = 32
	}

	if _, err := vbm(
		"modifyvm", name,
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
		return err
	}

	if _, err := vbm(
		"modifyvm", name,
		"--nic1", "nat",
		"--nictype1", "82540EM",
		"--cableconnected1", "on",
	); err != nil {
		return err
	}

	if _, err := vbm(
		"storagectl", name,
		"--name", "SATA",
		"--add", "sata",
		"--hostiocache", "on",
	); err != nil {
		return err
	}

	if _, err := vbm(
		"storageattach", name,
		"--storagectl", "SATA",
		"--port", "0",
		"--device", "0",
		"--type", "dvddrive",
		"--medium", filepath.Join(path, "boot2docker.iso"),
	); err != nil {
		return err
	}

	if _, err := vbm(
		"storageattach", name,
		"--storagectl", "SATA",
		"--port", "1",
		"--device", "0",
		"--type", "hdd",
		"--medium", filepath.Join(path, "disk.vmdk"),
	); err != nil {
		return err
	}

	if _, err := vbm(
		"guestproperty", "set", name,
		"/VirtualBox/GuestAdd/SharedFolders/MountPrefix", "/",
	); err != nil {
		return err
	}
	if _, err := vbm(
		"guestproperty", "set", name,
		"/VirtualBox/GuestAdd/SharedFolders/MountDir", "/",
	); err != nil {
		return err
	}

	shareName, shareDir := getShareDriveAndName()
	if _, err := os.Stat(shareDir); err != nil && !os.IsNotExist(err) {
		return err
	} else if !os.IsNotExist(err) {
		if shareName == "" {
			shareName = strings.TrimLeft(shareDir, "/")
		}

		if _, err := vbm(
			"sharedfolder", "add", name,
			"--name", shareName,
			"--hostpath", shareDir,
			"--automount",
		); err != nil {
			return err
		}

		if _, err := vbm(
			"setextradata", name,
			"VBoxInternal2/SharedFoldersEnableSymlinksCreate/"+shareName, "1",
		); err != nil {
			return err
		}
	}

	return nil
}

func Start(name string, storePath string) error {
	status, err := GetStatus(name)
	if err != nil {
		return err
	}

	sshForwarding := &PortForwarding{
		Name:      "ssh",
		Interface: 1,
		Protocol:  "tcp",
		GuestPort: 22,
	}

	switch status {
	case Paused:
		log.Info("resuming VM ...")
		if _, err := vbm("controlvm", name, "resume", "--type", "headless"); err != nil {
			return err
		}

	case Saved:
		log.Info("resuming VM ...")
		if err := ConfigurePortForwarding(name, sshForwarding); err != nil {
			return err
		}

		if _, err := vbm("startvm", name, "--type", "headless"); err != nil {
			return err
		}

	case Stopped:
		log.Info("check network to re-create if needed...")
		if err := SetupHostOnlyNetwork(name, "192.168.99.1/24"); err != nil {
			return err
		}

		if err := ConfigurePortForwarding(name, sshForwarding); err != nil {
			return err
		}

		if _, err := vbm("startvm", name, "--type", "headless"); err != nil {
			return err
		}

	default:
		return errors.New("VM not in startable state")
	}

	log.Info("waiting for an IP...")
	for i := 0; i < 60; i++ {
		if ip, err := GetIP(name, storePath); err == nil && ip != "" {
			return nil
		}
		time.Sleep(4 * time.Second)
	}

	return errors.New("could not get IP address")

}

func Stop(name string) error {
	status, err := GetStatus(name)
	if err != nil {
		return err
	}

	if status == Paused {
		if _, err := vbm("controlvm", name, "resume"); err != nil {
			return err
		}
		log.Info("resuming VM...")
	}

	if _, err := vbm("controlvm", name, "acpipowerbutton"); err != nil {
		return err
	}
	for {
		status, err = GetStatus(name)
		if err != nil {
			return err
		}
		if status == Running {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	return nil
}

func Remove(name string) error {
	status, err := GetStatus(name)
	if err != nil {
		return err
	}

	if status != Stopped && status != Saved {
		if _, err := vbm("controlvm", name, "poweroff"); err != nil {
			return err
		}
	}

	_, err = vbm("unregistervm", "--delete", name)
	return err
}

func GetURL(name string, storePath string) (string, error) {
	ip, err := GetIP(name, storePath)
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", errors.New("could not get IP")
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func GetIP(name string, storePath string) (string, error) {
	status, err := GetStatus(name)
	if err != nil {
		return "", err
	}
	if status != Running {
		return "", errors.New("VM is not running")
	}

	stdout, err := vbm("showvminfo", name, "--machinereadable")
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

	opts, err := GetSSHOptions(name)
	if err != nil {
		return "", err
	}

	executor, err := ssh.GimmeExecutor(&ssh.Options{
		Host:              opts.Hostname,
		Port:              opts.Port,
		User:              opts.Username,
		IdentityFileGlobs: []string{filepath.Join(storePath, "id_rsa")},
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
