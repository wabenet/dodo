package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v2"
)

func errorUnsupportedType(name string, kind reflect.Kind) error {
	return fmt.Errorf("Unsupported type of '%s': '%v'", name, kind)
}

func errorUnsupportedKey(parent string, name string) error {
	return fmt.Errorf("Unsupported option in '%s': '%s'", parent, name)
}

// ParseConfigurationFile reads a full dodo configuration from a file.
func ParseConfigurationFile(filename string) (Group, error) {
	if !filepath.IsAbs(filename) {
		directory, err := os.Getwd()
		if err != nil {
			return Group{}, err
		}
		filename, err = filepath.Abs(filepath.Join(directory, filename))
		if err != nil {
			return Group{}, err
		}
	}
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return Group{}, fmt.Errorf("Could not read file '%s'", filename)
	}
	return ParseConfiguration(filename, bytes)
}

// ParseConfiguration reads a full dodo configuration from text.
func ParseConfiguration(name string, bytes []byte) (Group, error) {
	var mapType map[interface{}]interface{}
	err := yaml.Unmarshal(bytes, &mapType)
	if err != nil {
		return Group{}, err
	}
	config, err := decodeGroup(name, mapType)
	if err != nil {
		return Group{}, err
	}
	return config, nil
}

func decodeIncludes(name string, config interface{}) ([]Group, error) {
	result := []Group{}
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

func decodeInclude(name string, config interface{}) (Group, error) {
	var result Group
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
