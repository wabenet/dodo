package config

import (
	"fmt"
	"reflect"

	"gopkg.in/yaml.v2"
)

// TODO: allow the following:
// - only the context as string in build
// - environment as key=value list or as map
// - volumes as source:dest:type list or as special structs
// - builds args as key=value list or as map
// - steps as string or slice

// TODO: support env_file as well

// Config represents a full configuration file
type Config struct {
	Backdrops Backdrops
}

func errorUnsupportedType(name string, kind reflect.Kind) error {
	return fmt.Errorf("Unsupported type of '%s': '%v'", name, kind)
}

func errorUnsupportedKey(parent string, name string) error {
	return fmt.Errorf("Unsupported option in '%s': '%s'", parent, name)
}

func ParseConfiguration(name string, bytes []byte) (*Config, error) {
	var mapType map[string]interface{}
	err := yaml.Unmarshal(bytes, &mapType)
	if err != nil {
		return nil, err
	}
	config, err := DecodeConfig(name, mapType)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func DecodeConfig(name string, config interface{}) (Config, error) {
	var result Config
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[string]interface{}) {
			switch k {
			case "backdrops":
				decoded, err := DecodeBackdrops(k, v)
				if err != nil {
					return result, err
				}
				result.Backdrops = decoded
			default:
				return result, errorUnsupportedKey(name, k)
			}
		}
	default:
		return result, errorUnsupportedType(name, t.Kind())
	}
	return result, nil
}
