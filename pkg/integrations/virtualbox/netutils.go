package virtualbox

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func parseIPv4Mask(s string) net.IPMask {
	mask := net.ParseIP(s)
	if mask == nil {
		return nil
	}
	return net.IPv4Mask(mask[12], mask[13], mask[14], mask[15])
}

func findAvailableTCPPort() (int, error) {
	for i := 0; i <= 10; i++ {
		ln, err := net.Listen("tcp4", "127.0.0.1:0")
		if err != nil {
			return 0, err
		}
		defer ln.Close()
		addr := ln.Addr().String()
		addrParts := strings.SplitN(addr, ":", 2)
		port, err := strconv.Atoi(addrParts[1])
		if err != nil {
			return 0, err
		}
		if port != 0 {
			return port, nil
		}
		time.Sleep(1 * time.Second)
	}
	return 0, errors.New("unable to allocate TCP port")
}
