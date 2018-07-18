package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v2"
)

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

// ParseConfigurationFile reads a full dodo configuration from a file.
func ParseConfigurationFile(filename string) (Config, error) {
	if !filepath.IsAbs(filename) {
		directory, err := os.Getwd()
		if err != nil {
			return Config{}, err
		}
		filename, err = filepath.Abs(filepath.Join(directory, filename))
		if err != nil {
			return Config{}, err
		}
	}
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return Config{}, fmt.Errorf("Could not read file '%s'", filename)
	}
	return ParseConfiguration(filename, bytes)
}

// ParseConfiguration reads a full dodo configuration from text.
func ParseConfiguration(name string, bytes []byte) (Config, error) {
	var mapType map[string]interface{}
	err := yaml.Unmarshal(bytes, &mapType)
	if err != nil {
		return Config{}, err
	}
	config, err := decodeConfig(name, mapType)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}

func decodeConfig(name string, config interface{}) (Config, error) {
	// TODO: this method kinda falls out of the normal schema
	backdrops := Backdrops{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[string]interface{}) {
			switch k {
			case "backdrops":
				decoded, err := decodeBackdrops(k, v)
				if err != nil {
					return Config{}, err
				}
				for name, backdrop := range decoded {
					backdrops[name] = backdrop
				}
			case "include":
				decoded, err := decodeIncludes(k, v)
				if err != nil {
					return Config{}, err
				}
				for _, include := range decoded {
					for name, backdrop := range include.Backdrops {
						backdrops[name] = backdrop
					}
				}
			default:
				return Config{}, errorUnsupportedKey(name, k)
			}
		}
	default:
		return Config{}, errorUnsupportedType(name, t.Kind())
	}
	return Config{Backdrops: backdrops}, nil
}

func decodeIncludes(name string, config interface{}) ([]Config, error) {
	result := []Config{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		decoded, err := decodeInclude(name, config)
		if err != nil {
			return result, err
		}
		result = append(result, decoded)
	case reflect.Slice:
		for _, v := range t.Interface().([]interface{}) {
			decoded, err := decodeInclude(name, v)
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

func decodeInclude(name string, config interface{}) (Config, error) {
	var result Config
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "file":
				decoded, err := decodeString(key, v)
				if err != nil {
					return result, err
				}
				return ParseConfigurationFile(decoded)
			case "text":
				decoded, err := decodeString(key, v)
				if err != nil {
					return result, err
				}
				return ParseConfiguration(name, []byte(decoded))
			default:
				return result, errorUnsupportedKey(name, key)
			}
		}
	default:
		return result, errorUnsupportedType(name, t.Kind())
	}
	return result, nil
}
