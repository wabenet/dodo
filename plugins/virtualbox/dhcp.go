package main

import (
	"net"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

const dhcpPrefix = "HostInterfaceNetworking-"

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

func ListDHCPServers() (map[string]*DHCPServer, error) {
	stdout, err := vbm("list", "dhcpservers")
	if err != nil {
		return nil, err
	}

	result := map[string]*DHCPServer{}
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
			result[groups[2]] = current
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

func CleanupDHCPServers() error {
	servers, err := ListDHCPServers()
	if err != nil {
		return err
	}

	if len(servers) == 0 {
		return nil
	}

	nets, err := ListHostOnlyAdapters()
	if err != nil {
		return err
	}

	for name := range servers {
		if strings.HasPrefix(name, dhcpPrefix) {
			if _, present := nets[name]; !present {
				if _, err := vbm("dhcpserver", "remove", "--netname", name); err != nil {
					log.WithFields(log.Fields{
						"server": name,
						"error":  err,
					}).Warn("could not remove dhcp server")
				}
			}
		}
	}

	return nil
}

func AddDHCPServer(ifname string, config DHCPServer) error {
	servers, err := ListDHCPServers()
	if err != nil {
		return err
	}

	name := dhcpPrefix + ifname
	current, exist := servers[name]

	if exist && current.Equal(config) {
		return nil
	}

	var command string
	if exist {
		command = "modify"
	} else {
		command = "add"
	}

	var enableFlag string
	if config.Enabled {
		enableFlag = "--enable"
	} else {
		enableFlag = "--disable"
	}

	_, err = vbm(
		"dhcpserver", command, enableFlag,
		"--netname", name,
		"--ip", config.IPv4.IP.String(),
		"--netmask", net.IP(config.IPv4.Mask).String(),
		"--lowerip", config.LowerIP.String(),
		"--upperip", config.UpperIP.String(),
	)
	return err
}
