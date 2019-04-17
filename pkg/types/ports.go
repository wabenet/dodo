package types

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/docker/go-connections/nat"
)

// Ports represents a set of port bindings
type Ports []Port

// Port represents a port binding
type Port struct {
	Target    string
	Published string
	Protocol  string
	HostIP    string
}

func (ports Ports) PortMap() nat.PortMap {
	result := map[nat.Port][]nat.PortBinding{}
	for _, port := range ports {
		portSpec, _ := nat.NewPort(port.Protocol, port.Target)
		result[portSpec] = append(result[portSpec], nat.PortBinding{HostPort: port.Published})
	}
	return result
}

func (ports Ports) PortSet() nat.PortSet {
	result := map[nat.Port]struct{}{}
	for _, port := range ports {
		portSpec, _ := nat.NewPort(port.Protocol, port.Target)
		result[portSpec] = struct{}{}
	}
	return result
}

// DecodePorts creates port binding configurations from a config map.
func DecodePorts(name string, config interface{}) (Ports, error) {
	result := []Port{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		decoded, err := DecodePort(name, config)
		if err != nil {
			return result, err
		}
		result = append(result, decoded)
	case reflect.Map:
		decoded, err := DecodePort(name, config)
		if err != nil {
			return result, err
		}
		result = append(result, decoded)
	case reflect.Slice:
		for _, v := range t.Interface().([]interface{}) {
			decoded, err := DecodePort(name, v)
			if err != nil {
				return result, err
			}
			result = append(result, decoded)
		}
	default:
		return result, ErrorUnsupportedType(name, t.Kind())
	}
	return result, nil
}

// DecodePort creates a port binding configuration from a config map.
func DecodePort(name string, config interface{}) (Port, error) {
	result := Port{Protocol: "tcp"}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		decoded, err := DecodeString(name, t.String())
		if err != nil {
			return result, err
		}
		switch values := strings.SplitN(decoded, ":", 3); len(values) {
		case 0:
			return result, fmt.Errorf("empty port definition in '%s'", name)
		case 1:
			result.Target = values[0]
		case 2:
			result.Published = values[0]
			result.Target = values[1]
		case 3:
			result.HostIP = values[0]
			result.Published = values[1]
			result.Target = values[2]
		default:
			return result, fmt.Errorf("too many values in '%s'", name)
		}
		switch values := strings.SplitN(result.Target, "/", 2); len(values) {
		case 1:
			result.Target = values[0]
		case 2:
			result.Target = values[0]
			result.Protocol = values[1]
		default:
			return result, fmt.Errorf("Too many values in '%s'", name)
		}
		return result, nil
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "target":
				decoded, err := DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Target = decoded
			case "published":
				decoded, err := DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Published = decoded
			case "protocol":
				decoded, err := DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Protocol = decoded
			case "host_ip":
				decoded, err := DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.HostIP = decoded
			default:
				return result, ErrorUnsupportedKey(name, key)
			}
		}
		return result, nil
	default:
		return result, ErrorUnsupportedType(name, t.Kind())
	}
}
