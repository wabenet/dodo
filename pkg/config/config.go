package config

import (
	"fmt"
	"reflect"

	"gopkg.in/yaml.v2"
)

// Config represents a full configuration file
type Config struct {
	Backdrops Backdrops
	Includes  Includes
}

func errorUnsupportedType(name string, kind reflect.Kind) error {
	return fmt.Errorf("Unsupported type of '%s': '%v'", name, kind)
}

func errorUnsupportedKey(parent string, name string) error {
	return fmt.Errorf("Unsupported option in '%s': '%s'", parent, name)
}

// ParseConfiguration reads a full dodo configuration from text.
func ParseConfiguration(name string, bytes []byte) (*Config, error) {
	var mapType map[string]interface{}
	err := yaml.Unmarshal(bytes, &mapType)
	if err != nil {
		return nil, err
	}
	config, err := decodeConfig(name, mapType)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func decodeConfig(name string, config interface{}) (Config, error) {
	var result Config
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[string]interface{}) {
			switch k {
			case "backdrops":
				decoded, err := decodeBackdrops(k, v)
				if err != nil {
					return result, err
				}
				result.Backdrops = decoded
			case "include":
				decoded, err := decodeIncludes(k, v)
				if err != nil {
					return result, err
				}
				result.Includes = decoded
			default:
				return result, errorUnsupportedKey(name, k)
			}
		}
	default:
		return result, errorUnsupportedType(name, t.Kind())
	}
	return result, nil
}
