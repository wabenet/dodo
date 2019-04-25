package types

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v2"
)

type Groups map[string]Group

type Group struct {
	Backdrops Backdrops
	Groups    Groups
}

func DecodeGroups(name string, config interface{}) (Groups, error) {
	result := map[string]Group{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			key := k.(string)
			decoded, err := DecodeGroup(key, v)
			if err != nil {
				return result, err
			}
			result[key] = decoded
		}
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}

func DecodeGroup(name string, config interface{}) (Group, error) {
	result := Group{Backdrops: Backdrops{}, Groups: Groups{}}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "groups":
				decoded, err := DecodeGroups(key, v)
				if err != nil {
					return result, err
				}
				for name, group := range decoded {
					result.Groups[name] = group
				}
			case "backdrops":
				decoded, err := DecodeBackdrops(key, v)
				if err != nil {
					return result, err
				}
				for name, backdrop := range decoded {
					result.Backdrops[name] = backdrop
				}
			case "include":
				decoded, err := DecodeIncludes(key, v)
				if err != nil {
					return result, err
				}
				for _, include := range decoded {
					for name, backdrop := range include.Backdrops {
						result.Backdrops[name] = backdrop
					}
				}
			default:
				return result, &ConfigError{Name: name, UnsupportedKey: &key}
			}
		}
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}

func DecodeIncludes(name string, config interface{}) ([]Group, error) {
	result := []Group{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		decoded, err := DecodeInclude(name, config)
		if err != nil {
			return result, err
		}
		result = append(result, decoded)
	case reflect.Slice:
		for _, v := range t.Interface().([]interface{}) {
			decoded, err := DecodeInclude(name, v)
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

func DecodeInclude(name string, config interface{}) (Group, error) {
	var result Group
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "file":
				decoded, err := DecodeString(key, v)
				if err != nil {
					return result, err
				}
				return includeFile(decoded)
			case "text":
				decoded, err := DecodeString(key, v)
				if err != nil {
					return result, err
				}
				return includeText(name, []byte(decoded))
			default:
				return result, &ConfigError{Name: name, UnsupportedKey: &key}
			}
		}
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}

func includeFile(filename string) (Group, error) {
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
		return Group{}, fmt.Errorf("could not read file '%s'", filename)
	}
	return includeText(filename, bytes)
}

func includeText(name string, bytes []byte) (Group, error) {
	var mapType map[interface{}]interface{}
	err := yaml.Unmarshal(bytes, &mapType)
	if err != nil {
		return Group{}, err
	}
	config, err := DecodeGroup(name, mapType)
	if err != nil {
		return Group{}, err
	}
	return config, nil
}
