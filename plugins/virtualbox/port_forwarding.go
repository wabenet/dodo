package main

import (
	"encoding/csv"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type PortForwarding struct {
	Interface int
	Name      string
	Protocol  string
	HostIP    string
	HostPort  int
	GuestIP   string
	GuestPort int
}

func (forward *PortForwarding) ToCSV() string {
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

func ListPortForwardings(name string) ([]PortForwarding, error) {
	var result []PortForwarding

	stdout, err := vbm("showvminfo", name, "--machinereadable")
	if err != nil {
		return result, errors.Wrap(err, "could not get machine info")
	}

	re := regexp.MustCompile(`(?m)^Forwarding\(([\d]+)\)="(.*)"`)
	for _, groups := range re.FindAllStringSubmatch(stdout, -1) {
		iface, err := strconv.Atoi(groups[1])
		if err != nil {
			return result, err
		}
		reader := csv.NewReader(strings.NewReader(groups[2]))
		fields, err := reader.Read()
		if err != nil {
			return result, errors.Wrap(err, "could not parse csv")
		}
		if len(fields) != 6 {
			return result, errors.New("unexpected numberf ouf fields")
		}
		hostPort, err := strconv.Atoi(fields[3])
		if err != nil {
			return result, err
		}
		guestPort, err := strconv.Atoi(fields[5])
		if err != nil {
			return result, err
		}
		result = append(result, PortForwarding{
			Interface: iface,
			Name:      fields[0],
			Protocol:  fields[1],
			HostIP:    fields[2],
			HostPort:  hostPort,
			GuestIP:   fields[4],
			GuestPort: guestPort,
		})
	}

	return result, nil
}

func ConfigurePortForwarding(name string, config *PortForwarding) error {
	hostPort, err := findAvailableTCPPort()
	if err != nil {
		return err
	}

	config.HostIP = "127.0.0.1"
	config.HostPort = hostPort

	vbm(
		"modifyvm", name,
		fmt.Sprintf("--natpf%d", config.Interface),
		"delete", config.Name,
	)

	_, err = vbm(
		"modifyvm", name,
		fmt.Sprintf("--natpf%d", config.Interface),
		config.ToCSV(),
	)
	return err
}

func findAvailableTCPPort() (int, error) {
	port := 0
	for i := 0; i <= 10; i++ {
		ln, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			return 0, err
		}
		defer ln.Close()
		addr := ln.Addr().String()
		addrParts := strings.SplitN(addr, ":", 2)
		p, err := strconv.Atoi(addrParts[1])
		if err != nil {
			return 0, err
		}
		if p != 0 {
			port = p
			return port, nil
		}
		port = 0
		time.Sleep(1 * time.Second)
	}
	return 0, fmt.Errorf("unable to allocate tcp port")
}
