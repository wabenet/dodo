package virtualbox

import (
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var empty interface{}

type HostOnlyNetwork struct {
	Name        string
	DHCP        bool
	IPv4        net.IPNet
	NetworkName string
}

func NewHostOnlyNetwork(cidr string) (*HostOnlyNetwork, error) {
	ip, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	result := &HostOnlyNetwork{IPv4: *network}
	result.IPv4.IP = ip

	return result, nil
}

func (network HostOnlyNetwork) Equal(other HostOnlyNetwork) bool {
	if !network.IPv4.IP.Equal(other.IPv4.IP) {
		return false
	}

	if network.IPv4.Mask.String() != other.IPv4.Mask.String() {
		if network.IPv4.Mask.String() != "0f000000" {
			return false
		}
	}

	if network.DHCP != other.DHCP {
		return false
	}

	return true
}

func (network *HostOnlyNetwork) Create() error {
	existingNetworks, err := ListHostOnlyNetworks()
	if err != nil {
		return err
	}

	networkNames := map[string]interface{}{}
	for _, n := range existingNetworks {
		if n.Equal(*network) {
			network.Name = n.Name
			network.NetworkName = n.NetworkName
			return nil
		}
		networkNames[n.NetworkName] = empty
	}

	vbm("hostonlyif", "create")

	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)

		currentNetworks, err := ListHostOnlyNetworks()
		if err != nil {
			return err
		}

		for _, newNetwork := range currentNetworks {
			if _, ok := networkNames[newNetwork.NetworkName]; ok {
				continue
			}

			network.Name = newNetwork.Name
			network.NetworkName = newNetwork.NetworkName

			if _, err := vbm(
				"hostonlyif", "ipconfig", network.Name,
				"--ip", network.IPv4.IP.String(),
				"--netmask", net.IP(network.IPv4.Mask).String(),
			); err != nil {
				return err
			}
			if network.DHCP {
				vbm("hostonlyif", "ipconfig", network.Name, "--dhcp")
			}

			return nil
		}
	}

	return errors.New("could not find a new host-only adapter")
}

func (network *HostOnlyNetwork) ConnectVM(vm *VM) error {
	return vm.Modify(
		"--nic2", "hostonly",
		"--nictype2", "82540EM",
		"--nicpromisc2", "deny",
		"--hostonlyadapter2", network.Name,
		"--cableconnected2", "on",
	)
}

func ListHostOnlyNetworks() ([]*HostOnlyNetwork, error) {
	stdout, err := vbm("list", "hostonlyifs")
	if err != nil {
		return nil, err
	}

	result := []*HostOnlyNetwork{}
	current := &HostOnlyNetwork{}
	re := regexp.MustCompile(`(.+):\s+(.*)`)
	for _, line := range strings.Split(stdout, "\n") {
		if line == "" {
			continue
		}

		groups := re.FindStringSubmatch(line)
		if groups == nil {
			continue
		}

		switch groups[1] {
		case "Name":
			current = &HostOnlyNetwork{Name: groups[2]}
			result = append(result, current)
		case "DHCP":
			current.DHCP = (groups[2] != "Disabled")
		case "IPAddress":
			current.IPv4.IP = net.ParseIP(groups[2])
		case "NetworkMask":
			current.IPv4.Mask = parseIPv4Mask(groups[2])
		case "VBoxNetworkName":
			current.NetworkName = groups[2]
		}
	}

	return result, nil
}
