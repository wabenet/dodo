package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/oclaussen/dodo/pkg/types"
)

func decodeVolumes(name string, config interface{}) (types.Volumes, error) {
	result := []types.Volume{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		decoded, err := decodeVolume(name, config)
		if err != nil {
			return result, err
		}
		result = append(result, decoded)
	case reflect.Map:
		decoded, err := decodeVolume(name, config)
		if err != nil {
			return result, err
		}
		result = append(result, decoded)
	case reflect.Slice:
		for _, v := range t.Interface().([]interface{}) {
			decoded, err := decodeVolume(name, v)
			if err != nil {
				return result, err
			}
			result = append(result, decoded)
		}
	default:
		return result, errorUnsupportedType(name, t.Kind())
	}
	return result, nil
}

func decodeVolume(name string, config interface{}) (types.Volume, error) {
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		decoded, err := decodeString(name, t.String())
		if err != nil {
			return types.Volume{}, err
		}
		switch values := strings.SplitN(decoded, ":", 3); len(values) {
		case 0:
			return types.Volume{}, fmt.Errorf("Empty volume definition in '%s'", name)
		case 1:
			return types.Volume{
				Source: values[0],
			}, nil
		case 2:
			return types.Volume{
				Source: values[0],
				Target: values[1],
			}, nil
		case 3:
			return types.Volume{
				Source:   values[0],
				Target:   values[1],
				ReadOnly: values[2] == "ro",
			}, nil
		default:
			return types.Volume{}, fmt.Errorf("Too many values in '%s'", name)
		}
	case reflect.Map:
		var result types.Volume
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "source":
				decoded, err := decodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Source = decoded
			case "target":
				decoded, err := decodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Target = decoded
			case "read_only":
				decoded, err := decodeBool(key, v)
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
		return types.Volume{}, errorUnsupportedType(name, t.Kind())
	}
}
