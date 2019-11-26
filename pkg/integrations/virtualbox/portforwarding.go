package virtualbox

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type PortForwarding struct {
	VMName    string
	Interface int
	Name      string
	Protocol  string
	HostIP    string
	HostPort  int
	GuestIP   string
	GuestPort int
}

func (vm *VM) NewPortForwarding(name string) *PortForwarding {
	return &PortForwarding{
		VMName:    vm.Name,
		Interface: 1,
		Name:      name,
		Protocol:  "tcp",
		HostIP:    "127.0.0.1",
		HostPort:  0,
		GuestIP:   "",
		GuestPort: 0,
	}
}

func fromCSV(vmname string, iface int, csvString string) (*PortForwarding, error) {
	reader := csv.NewReader(strings.NewReader(csvString))
	fields, err := reader.Read()
	if err != nil {
		return nil, errors.Wrap(err, "could not parse csv")
	}
	if len(fields) != 6 {
		return nil, errors.New("unexpected number of fields")
	}

	hostPort, err := strconv.Atoi(fields[3])
	if err != nil {
		return nil, err
	}

	guestPort, err := strconv.Atoi(fields[5])
	if err != nil {
		return nil, err
	}

	return &PortForwarding{
		VMName:    vmname,
		Interface: iface,
		Name:      fields[0],
		Protocol:  fields[1],
		HostIP:    fields[2],
		HostPort:  hostPort,
		GuestIP:   fields[4],
		GuestPort: guestPort,
	}, nil
}

func (forward *PortForwarding) toCSV() string {
	return fmt.Sprintf(
		"%s,%s,%s,%d,%s,%d",
		forward.Name,
		forward.Protocol,
		forward.HostIP,
		forward.HostPort,
		forward.GuestIP,
		forward.GuestPort,
	)
}

func (forward *PortForwarding) Create() error {
	if forward.HostPort == 0 {
		port, err := findAvailableTCPPort()
		if err != nil {
			return err
		}
		forward.HostPort = port
	}

	vbm(
		"modifyvm", forward.VMName,
		fmt.Sprintf("--natpf%d", forward.Interface),
		"delete", forward.Name,
	)

	_, err := vbm(
		"modifyvm", forward.VMName,
		fmt.Sprintf("--natpf%d", forward.Interface),
		forward.toCSV(),
	)
	return err
}

func (vm *VM) ListPortForwardings() ([]*PortForwarding, error) {
	var result []*PortForwarding

	info, err := vm.Info()
	if err != nil {
		return result, err
	}

	for name, value := range info {
		if !strings.HasPrefix(name, "Forwarding") {
			continue
		}
		name = strings.TrimPrefix(name, "Forwarding(")
		name = strings.TrimSuffix(name, ")")
		iface, err := strconv.Atoi(name)
		if err != nil {
			return result, err
		}
		forward, err := fromCSV(vm.Name, iface, value)
		if err != nil {
			return result, err
		}
		result = append(result, forward)
	}

	return result, nil
}
