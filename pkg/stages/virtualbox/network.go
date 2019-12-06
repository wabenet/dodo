package virtualbox

import (
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/oclaussen/dodo/pkg/integrations/virtualbox"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	dhcpPrefix = "HostInterfaceNetworking-"
)

func (vbox *Stage) SetupHostOnlyNetwork(cidr string) error {
	ip, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}

	err = CheckNetworkCollisions(network)
	if err != nil {
		return err
	}

	hostOnlyNetwork, err := virtualbox.NewHostOnlyNetwork(cidr)
	if err != nil {
		return err
	}
	if err := hostOnlyNetwork.Create(); err != nil {
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

	server := virtualbox.DHCPServer{
		NetworkName: dhcpPrefix + hostOnlyNetwork.Name,
		IPv4:        net.IPNet{IP: dhcpAddr, Mask: network.Mask},
		LowerIP:     lowerIP,
		UpperIP:     upperIP,
		Enabled:     true,
	}
	if err := server.Create(); err != nil {
		return err
	}

	if err := hostOnlyNetwork.ConnectVM(vbox.VM); err != nil {
		return err
	}

	return nil
}

func CheckNetworkCollisions(target *net.IPNet) error {
	hostonlyifs, err := virtualbox.ListHostOnlyNetworks()
	if err != nil {
		return err
	}

	var dummy interface{}
	excludedNetworks := map[string]interface{}{}
	for _, network := range hostonlyifs {
		ipnet := net.IPNet{IP: network.IPv4.IP, Mask: network.IPv4.Mask}
		excludedNetworks[ipnet.String()] = dummy
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return err
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
				if target.IP.Equal(ipnet.IP.Mask(ipnet.Mask)) {
					return errors.New("target cidr conflicts with the network address of a host interface")
				}
			}
		}
	}

	return nil
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

func CleanupDHCPServers() error {
	servers, err := virtualbox.ListDHCPServers()
	if err != nil {
		return err
	}
	if len(servers) == 0 {
		return nil
	}

	hostonlyifs, err := virtualbox.ListHostOnlyNetworks()
	if err != nil {
		return err
	}

	var dummy interface{}
	currentNetworks := map[string]interface{}{}
	for _, network := range hostonlyifs {
		currentNetworks[network.NetworkName] = dummy
	}

	for _, server := range servers {
		if strings.HasPrefix(server.NetworkName, dhcpPrefix) {
			if _, present := currentNetworks[server.NetworkName]; !present {
				if err := server.Remove(); err != nil {
					log.WithFields(log.Fields{
						"server": server.NetworkName,
						"error":  err,
					}).Warn("could not remove dhcp server")
				}
			}
		}
	}

	return nil
}
