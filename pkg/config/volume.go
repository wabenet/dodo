package config

import (
	"fmt"
	"reflect"
	"strings"
)

// Volumes represents a set of volume configurations
type Volumes []Volume

// Volume represents a bind-mount volume configuration
type Volume struct {
	Source   string
	Target   string
	ReadOnly bool
}

func (vs *Volumes) Strings() []string {
	result := []string{}
	for _, v := range *vs {
		result = append(result, v.String())
	}
	return result
}

func (v *Volume) String() string {
	if v.Target == "" && !v.ReadOnly {
		return v.Source
	} else if !v.ReadOnly {
		return fmt.Sprintf("%s:%s", v.Source, v.Target)
	} else {
		return fmt.Sprintf("%s:%s:ro", v.Source, v.Target)
	}
}

func DecodeVolumes(name string, config interface{}) (Volumes, error) {
	result := []Volume{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		decoded, err := DecodeVolume(name, config)
		if err != nil {
			return result, err
		}
		result = append(result, decoded)
	case reflect.Map:
		decoded, err := DecodeVolume(name, config)
		if err != nil {
			return result, err
		}
		result = append(result, decoded)
	case reflect.Slice:
		for _, v := range t.Interface().([]interface{}) {
			decoded, err := DecodeVolume(name, v)
			if err != nil {
				return result, err
			}
			result = append(result, decoded)
		}
	}
	return result, nil
}

func DecodeVolume(name string, config interface{}) (Volume, error) {
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		switch values := strings.SplitN(t.String(), ":", 3); len(values) {
		case 0:
			return Volume{}, fmt.Errorf("Empty volume definition in '%s'", name)
		case 1:
			return Volume{
				Source: values[0],
			}, nil
		case 2:
			return Volume{
				Source: values[0],
				Target: values[1],
			}, nil
		case 3:
			return Volume{
				Source:   values[0],
				Target:   values[1],
				ReadOnly: values[2] == "ro",
			}, nil
		default:
			return Volume{}, fmt.Errorf("Too many values in '%s'", name)
		}
	case reflect.Map:
		var result Volume
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "source":
				decoded, err := DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Source = decoded
			case "target":
				decoded, err := DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Target = decoded
			case "read_only":
				decoded, err := DecodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.ReadOnly = decoded
			default:
				return result, errorUnsupportedKey(name, key)
			}
		}
		return result, nil
	default:
		return Volume{}, errorUnsupportedType(name, t.Kind())
	}
}
