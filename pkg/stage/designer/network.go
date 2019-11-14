package designer

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

const (
	dodoBegin = "##_DODO_BEGIN_##"
	dodoEnd   = "##_DODO_END_##"

	interfacesFile = "/etc/network/interfaces"
	networkScripts = "/etc/sysconfig/network-scripts"

	networkManagerConfig = `
DEVICE={{ .Device }}
ONBOOT=yes
BOOTPROTO=dhcp
`

	networkToolsConfig = `
auto {{ .Device }}
iface {{ .Device }} inet dhcp
post-up ip route del default dev $IFACE || true
`
)

type Network struct {
	Device string
}

func ConfigureNetwork(network Network) (string, error) {
	if _, err := exec.LookPath("nmcli"); err == nil {
		if err := configureNetworkManager(network); err != nil {
			return "", err
		}
	} else {
		if err := configureNetTools(network); err != nil {
			return "", err
		}
	}
	return getIP(network)
}

func configureNetworkManager(network Network) error {
	var buffer bytes.Buffer
	templ, err := template.New("network").Parse(networkManagerConfig)
	if err != nil {
		return err
	}
	if err := templ.Execute(&buffer, network); err != nil {
		return err
	}
	configFile := filepath.Join(networkScripts, fmt.Sprintf("ifcfg-%s", network.Device))
	if err := ioutil.WriteFile(configFile, buffer.Bytes(), 0644); err != nil {
		return err
	}

	if nmcli, err := exec.LookPath("nmcli"); err == nil {
		exec.Command(nmcli, "d", "disconnect", network.Device).Run()
	} else {
		exec.Command("/sbin/ifdown", network.Device).Run()
	}

	if systemctl, err := exec.LookPath("systemctl"); err == nil {
		if err := exec.Command(systemctl, "restart", "NetworkManager").Run(); err != nil {
			return errors.Wrap(err, "could not restart NetworkManager")
		}
	} else {
		if err := exec.Command("/etc/init.d/NetworkManager", "restart").Run(); err != nil {
			return errors.Wrap(err, "could not restart NetworkManager")
		}
	}
	return nil
}

func configureNetTools(network Network) error {
	// Do not error check because the device might not exist
	exec.Command("/sbin/ifdown", network.Device).Run()
	exec.Command("/sbin/ip", "addr", "flush", "dev", network.Device).Run()

	templ, err := template.New("network").Parse(networkToolsConfig)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	err = templ.Execute(&buffer, network)
	if err != nil {
		return err
	}

	if err := replaceBlockInFile(interfacesFile, dodoBegin, dodoEnd, buffer.String()); err != nil {
		return err
	}

	if err := exec.Command("/sbin/ifup", network.Device).Run(); err != nil {
		return err
	}

	return nil
}

func readSurroundingLines(path string, beginMarker string, endMarker string) ([]string, []string, error) {
	var preLines, postLines []string

	file, err := os.Open(path)
	if err != nil {
		return preLines, postLines, errors.Wrap(err, "could not open file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == beginMarker {
			break
		}
		preLines = append(preLines, line)
	}
	for scanner.Scan() {
		line := scanner.Text()
		if line == endMarker {
			break
		}
	}
	for scanner.Scan() {
		line := scanner.Text()
		postLines = append(postLines, line)
	}

	return preLines, postLines, scanner.Err()
}

func replaceBlockInFile(path string, beginMarker string, endMarker string, contents string) error {
	preLines, postLines, err := readSurroundingLines(path, beginMarker, endMarker)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "could not open file")
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range preLines {
		fmt.Fprintln(writer, line)
	}
	fmt.Fprintln(writer, beginMarker)
	fmt.Fprintln(writer, contents)
	fmt.Fprintln(writer, endMarker)
	for _, line := range postLines {
		fmt.Fprintln(writer, line)
	}
	if err := writer.Flush(); err != nil {
		return err
	}

	return nil
}

func getIP(network Network) (string, error) {
	for i := 0; i < 5; i++ {
		iface, err := net.InterfaceByName(network.Device)
		if err != nil {
			return "", err
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			if a, ok := addr.(*net.IPNet); ok {
				if ip := a.IP.To4(); ip != nil {
					return ip.String(), nil
				}
			}
		}
		time.Sleep(time.Second)
	}
	return "", errors.New("could not get host-only IP address")
}
