package virtualbox

import (
	"fmt"
	"strconv"
)

// TODO: support other medium types than hdd

const (
	scprefix = "storagecontroller"

	IDE    = "ide"
	SATA   = "sata"
	SCSI   = "scsi"
	SAS    = "sas"
	PCIE   = "pcie"
	Floppy = "floppy"
	USB    = "usb"
)

type Disk struct {
	Path string
	Size int64
}

type StorageController struct {
	VMName    string
	Name      string
	Type      string
	Device    int
	PortCount int
	Disks     []Disk
}

func (disk *Disk) Create() error {
	sizeInMB := disk.Size / 1000 / 1000
	_, err := vbm(
		"createhd",
		"--filename", disk.Path,
		"--size", strconv.FormatInt(sizeInMB, 10),
	)
	return err
}

func (vm *VM) ListStorageControllers() ([]*StorageController, error) {
	result := []*StorageController{}

	info, err := vm.Info()
	if err != nil {
		return result, err
	}

	i := 0
	for {
		name, ok := info[fmt.Sprintf("%sname%d", scprefix, i)]
		if !ok {
			break
		}

		ctl := &StorageController{VMName: vm.Name, Name: name}

		if count, ok := info[fmt.Sprintf("%sinstance%d", scprefix, i)]; ok {
			ctl.Device, _ = strconv.Atoi(count)
		}

		if count, ok := info[fmt.Sprintf("%sportcount%d", scprefix, i)]; ok {
			ctl.PortCount, _ = strconv.Atoi(count)
		}

		if t, ok := info[fmt.Sprintf("%stype%d", scprefix, i)]; ok {
			switch t {
			case "PIIX3", "PIIX4", "ICH6":
				ctl.Type = IDE
			case "IntelAhci":
				ctl.Type = SATA
			case "LsiLogic", "BusLogic":
				ctl.Type = SCSI
			case "LsiLogicSas":
				ctl.Type = SAS
			case "NVMe":
				ctl.Type = PCIE
			case "I82078":
				ctl.Type = Floppy
			case "USB":
				ctl.Type = USB
			}
		}

		ctl.Disks = make([]Disk, ctl.PortCount)
		for port := 0; port < ctl.PortCount; port++ {
			if path, ok := info[fmt.Sprintf("%s-%d-%d", ctl.Name, ctl.Device, port)]; ok {
				ctl.Disks[port] = Disk{Path: path}
			}
		}

		result = append(result, ctl)
		i++
	}

	return result, nil
}

func (vm *VM) GetStorageController(busType string) (*StorageController, error) {
	controllers, err := vm.ListStorageControllers()
	if err != nil {
		return nil, err
	}

	for _, ctl := range controllers {
		if ctl.Type == busType {
			return ctl, err
		}
	}

	ctl := &StorageController{
		VMName: vm.Name,
		Type:   busType,
	}

	if err := ctl.Create(); err != nil {
		return nil, err
	}

	return ctl, nil
}

func (ctl *StorageController) Create() error {
	if _, err := vbm(
		"storagectl", ctl.VMName,
		"--add", ctl.Type,
		"--name", ctl.Name,
		"--portcount", strconv.Itoa(ctl.PortCount),
	); err != nil {
		return err
	}

	for port, disk := range ctl.Disks {
		if err := ctl.AttachDisk(port, &disk); err != nil {
			return err
		}
	}

	return nil
}

func (ctl *StorageController) Remove() error {
	_, err := vbm("storagectl", ctl.VMName, "--remove", "--name", ctl.Name)
	return err
}

func (ctl *StorageController) AttachDisk(port int, disk *Disk) error {
	if ctl.PortCount <= len(ctl.Disks) {
		if err := ctl.Remove(); err != nil {
			return err
		}
		ctl.PortCount += 1
		if err := ctl.Create(); err != nil {
			return err
		}
		return ctl.AttachDisk(port, disk)
	}

	_, err := vbm(
		"storageattach", ctl.VMName,
		"--storagectl", ctl.Name,
		"--port", strconv.Itoa(port),
		"--device", strconv.Itoa(ctl.Device),
		"--type", "hdd",
		"--medium", disk.Path,
	)
	return err
}
