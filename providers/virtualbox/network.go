package main

import (
	"math/rand"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	buggyNetmask = "0f000000"
)

type hostOnlyNetwork struct {
	Name        string
	DHCP        bool
	IPv4        net.IPNet
	NetworkName string // referenced in DHCP.NetworkName
}

func ListHostOnlyAdapters() (map[string]*hostOnlyNetwork, error) {
	stdout, err := vbm("list", "hostonlyifs")
	if err != nil {
		return nil, err
	}

	result := map[string]*hostOnlyNetwork{}
	current := &hostOnlyNetwork{}
	re := regexp.MustCompile("(.+):\\s+(.*)")
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
			current.Name = groups[2]
		case "DHCP":
			current.DHCP = (groups[2] != "Disabled")
		case "IPAddress":
			current.IPv4.IP = net.ParseIP(groups[2])
		case "NetworkMask":
			current.IPv4.Mask = parseIPv4Mask(groups[2])
		case "VBoxNetworkName":
			current.NetworkName = groups[2]
			if _, present := result[current.NetworkName]; present {
				return result, errors.New("multiple host-only adapters with the same name exist")
			}
			result[current.NetworkName] = current
			current = &hostOnlyNetwork{}
		}
	}

	return result, nil
}

func SetupHostOnlyNetwork(machineName string, cidr string) error {
	ip, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}

	nets, err := ListHostOnlyAdapters()
	if err != nil {
		return err
	}

	err = CheckNetworkCollisions(network, nets)
	if err != nil {
		return err
	}

	hostOnlyAdapter, err := GetHostOnlyNetwork(ip, network.Mask, nets)
	if err != nil {
		return err
	}

	if err := CleanupDHCPServers(); err != nil {
		return err
	}

	dhcpAddr, err := getRandomIPinSubnet(ip)
	if err != nil {
		return err
	}

	lowerIP, upperIP := getDHCPAddressRange(dhcpAddr, network)

	dhcp := DHCPServer{}
	dhcp.IPv4.IP = dhcpAddr
	dhcp.IPv4.Mask = network.Mask
	dhcp.LowerIP = lowerIP
	dhcp.UpperIP = upperIP
	dhcp.Enabled = true
	if err := AddDHCPServer(hostOnlyAdapter.Name, dhcp); err != nil {
		return err
	}

	if _, err := vbm(
		"modifyvm", machineName,
		"--nic2", "hostonly",
		"--nictype2", "82540EM",
		"--nicpromisc2", "deny",
		"--hostonlyadapter2", hostOnlyAdapter.Name,
		"--cableconnected2", "on",
	); err != nil {
		return err
	}

	return nil
}

func GetHostOnlyNetwork(hostIP net.IP, netmask net.IPMask, nets map[string]*hostOnlyNetwork) (*hostOnlyNetwork, error) {
	for _, n := range nets {
		if hostIP.Equal(n.IPv4.IP) && (netmask.String() == n.IPv4.Mask.String() || n.IPv4.Mask.String() == buggyNetmask) {
			return n, nil
		}
	}

	vbm("hostonlyif", "create")

	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)

		newNets, err := ListHostOnlyAdapters()
		if err != nil {
			return nil, err
		}

		for name, latestNet := range newNets {
			if _, present := nets[name]; !present {

				latestNet.IPv4.IP = hostIP
				latestNet.IPv4.Mask = netmask

				if _, err := vbm(
					"hostonlyif", "ipconfig", latestNet.Name,
					"--ip", latestNet.IPv4.IP.String(),
					"--netmask", net.IP(latestNet.IPv4.Mask).String(),
				); err != nil {
					return nil, err
				}

				if latestNet.DHCP {
					vbm("hostonlyif", "ipconfig", latestNet.Name, "--dhcp")
				}

				return latestNet, nil
			}
		}
	}

	return nil, errors.New("failed to find a new host-only adapter")
}

func CheckNetworkCollisions(hostOnlyNet *net.IPNet, currentNetworks map[string]*hostOnlyNetwork) error {
	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	excludedNetworks := map[string]*hostOnlyNetwork{}
	for _, network := range currentNetworks {
		ipnet := net.IPNet{IP: network.IPv4.IP, Mask: network.IPv4.Mask}
		excludedNetworks[ipnet.String()] = network
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return err
		}

		excluded := false
		for _, addr := range addrs {
			switch ipnet := addr.(type) {
			case *net.IPNet:
				_, excluded = excludedNetworks[ipnet.String()]
				if excluded {
					break
				}
			}
		}

		if excluded || iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		for _, addr := range addrs {
			switch ipnet := addr.(type) {
			case *net.IPNet:
				if hostOnlyNet.IP.Equal(ipnet.IP.Mask(ipnet.Mask)) {
					return errors.New("host-only cidr conflicts with the network address of a host interface")
				}
			}
		}
	}

	return nil
}

func parseIPv4Mask(s string) net.IPMask {
	mask := net.ParseIP(s)
	if mask == nil {
		return nil
	}
	return net.IPv4Mask(mask[12], mask[13], mask[14], mask[15])
}

func getDHCPAddressRange(dhcpAddr net.IP, network *net.IPNet) (lowerIP net.IP, upperIP net.IP) {
	nAddr := network.IP.To4()
	ones, bits := network.Mask.Size()

	if ones <= 24 {
		// For a /24 subnet, use the original behavior of allowing the address range
		// between x.x.x.100 and x.x.x.254.
		lowerIP = net.IPv4(nAddr[0], nAddr[1], nAddr[2], byte(100))
		upperIP = net.IPv4(nAddr[0], nAddr[1], nAddr[2], byte(254))
		return
	}

	// Start the lowerIP range one address above the selected DHCP address.
	lowerIP = net.IPv4(nAddr[0], nAddr[1], nAddr[2], dhcpAddr.To4()[3]+1)

	// The highest D-part of the address A.B.C.D in this subnet is at 2^n - 1,
	// where n is the number of available bits in the subnet. Since the highest
	// address is reserved for subnet broadcast, the highest *assignable* address
	// is at (2^n - 1) - 1 == 2^n - 2.
	maxAssignableSubnetAddress := (byte)((1 << (uint)(bits-ones)) - 2)
	upperIP = net.IPv4(nAddr[0], nAddr[1], nAddr[2], maxAssignableSubnetAddress)
	return
}

func getRandomIPinSubnet(baseIP net.IP) (net.IP, error) {
	var dhcpAddr net.IP

	source := rand.New(rand.NewSource(time.Now().UnixNano()))

	nAddr := baseIP.To4()
	for i := 0; i < 5; i++ {
		n := source.Intn(24) + 1
		if byte(n) != nAddr[3] {
			dhcpAddr = net.IPv4(nAddr[0], nAddr[1], nAddr[2], byte(n))
			break
		}
	}

	if dhcpAddr == nil {
		return nil, errors.New("unable to generate random IP")
	}

	return dhcpAddr, nil
}
