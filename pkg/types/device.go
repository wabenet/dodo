package types

import (
	"fmt"
	"reflect"
	"strings"
)

type Devices []Device

type Device struct {
	CgroupRule  string
	Source      string
	Target      string
	Permissions string
}

func (d *decoder) DecodeDevices(name string, config interface{}) (Devices, error) {
	result := []Device{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String, reflect.Map:
		decoded, err := d.DecodeDevice(name, config)
		if err != nil {
			return result, err
		}
		result = append(result, decoded)
	case reflect.Slice:
		for _, v := range t.Interface().([]interface{}) {
			decoded, err := d.DecodeDevice(name, v)
			if err != nil {
				return result, err
			}
			result = append(result, decoded)
		}
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}

func (d *decoder) DecodeDevice(name string, config interface{}) (Device, error) {
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		decoded, err := d.DecodeString(name, t.String())
		if err != nil {
			return Device{}, err
		}
		switch values := strings.SplitN(decoded, ":", 3); len(values) {
		case 0:
			return Device{}, fmt.Errorf("empty volume definition in '%s'", name)
		case 1:
			return Device{
				Source: values[0],
			}, nil
		case 2:
			return Device{
				Source: values[0],
				Target: values[1],
			}, nil
		case 3:
			return Device{
				Source:      values[0],
				Target:      values[1],
				Permissions: values[2],
			}, nil
		default:
			return Device{}, fmt.Errorf("too many values in '%s'", name)
		}
	case reflect.Map:
		var result Device
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "cgroup_rule":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.CgroupRule = decoded
			case "source":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Source = decoded
			case "target":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Target = decoded
			case "read_only":
				decoded, err := d.DecodeBool(key, v)
				if err != nil {
					return result, err
				}
				if decoded {
					result.Permissions = "ro"
				}
			default:
				return result, &ConfigError{Name: name, UnsupportedKey: &key}
			}
		}
		return result, nil
	default:
		return Device{}, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
}
