package virtualbox

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

type DHCPServer struct {
	NetworkName string
	IPv4        net.IPNet
	LowerIP     net.IP
	UpperIP     net.IP
	Enabled     bool
}

func (server DHCPServer) Equal(other DHCPServer) bool {
	if !server.IPv4.IP.Equal(other.IPv4.IP) {
		return false
	}
	if server.IPv4.Mask.String() != other.IPv4.Mask.String() {
		return false
	}
	if !server.LowerIP.Equal(other.LowerIP) {
		return false
	}
	if !server.UpperIP.Equal(other.UpperIP) {
		return false
	}
	if server.Enabled != other.Enabled {
		return false
	}
	return true
}

func (server *DHCPServer) Create() error {
	current, err := GetDHCPServer(server.NetworkName)
	if err == nil && current != nil && server.Equal(*current) {
		return nil
	}

	args := []string{"dhcpserver"}

	if err == nil && current != nil {
		args = append(args, "modify")
	} else {
		args = append(args, "add")
	}
	if server.Enabled {
		args = append(args, "--enable")
	} else {
		args = append(args, "--disable")
	}

	args = append(args,
		"--netname", server.NetworkName,
		"--ip", server.IPv4.IP.String(),
		"--netmask", net.IP(server.IPv4.Mask).String(),
		"--lowerip", server.LowerIP.String(),
		"--upperip", server.UpperIP.String(),
	)

	_, err = vbm(args...)
	return err
}

func (server *DHCPServer) Remove() error {
	_, err := vbm("dhcpserver", "remove", "--netname", server.NetworkName)
	return err
}

func ListDHCPServers() ([]*DHCPServer, error) {
	stdout, err := vbm("list", "dhcpservers")
	if err != nil {
		return nil, err
	}

	var result []*DHCPServer
	current := &DHCPServer{}
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
		case "NetworkName":
			current = &DHCPServer{NetworkName: groups[2]}
			result = append(result, current)
		case "IP":
			current.IPv4.IP = net.ParseIP(groups[2])
		case "upperIPAddress":
			current.UpperIP = net.ParseIP(groups[2])
		case "lowerIPAddress":
			current.LowerIP = net.ParseIP(groups[2])
		case "NetworkMask":
			current.IPv4.Mask = parseIPv4Mask(groups[2])
		case "Enabled":
			current.Enabled = (groups[2] == "Yes")
		}
	}

	return result, nil
}

func GetDHCPServer(networkName string) (*DHCPServer, error) {
	servers, err := ListDHCPServers()
	if err != nil {
		return nil, err
	}
	for _, server := range servers {
		if server.NetworkName == networkName {
			return server, nil
		}
	}
	return nil, fmt.Errorf("dhcp server for network %s does not exist", networkName)
}
