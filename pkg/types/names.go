package types

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v2"
)

type Names struct {
	Backdrops map[string]string
	Groups    map[string]Names
}

func (names *Names) Names() []string {
	var result []string
	if names.Backdrops != nil {
		for name, _ := range names.Backdrops {
			result = append(result, name)
		}
	}
	if names.Groups != nil {
		for _, group := range names.Groups {
			result = append(result, group.Names()...)
		}
	}
	return result
}

func (names *Names) Strings() []string {
	var result []string
	if names.Backdrops != nil {
		for name, path := range names.Backdrops {
			result = append(result, fmt.Sprintf("%s (%s)", name, path))
		}
	}
	if names.Groups != nil {
		for name, group := range names.Groups {
			for _, substring := range group.Strings() {
				result = append(result, fmt.Sprintf("%s/%s", name, substring))
			}
		}
	}
	return result
}

func (target *Names) Merge(source *Names) {
	if source.Backdrops != nil {
		if target.Backdrops == nil {
			target.Backdrops = map[string]string{}
		}
		for key, value := range source.Backdrops {
			if _, ok := target.Backdrops[key]; !ok {
				target.Backdrops[key] = value
			}
		}
	}

	if source.Groups != nil {
		if target.Groups == nil {
			target.Groups = map[string]Names{}
		}
		for key, value := range source.Groups {
			if group, ok := target.Groups[key]; !ok {
				target.Groups[key] = value
			} else {
				group.Merge(&value)
			}
		}
	}
}

func (d *decoder) DecodeNames(path string, name string, config interface{}) (Names, error) {
	result := Names{Backdrops: map[string]string{}, Groups: map[string]Names{}}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "groups":
				decoded, err := d.DecodeNamesGroups(path, key, v)
				if err != nil {
					return result, err
				}
				for name, group := range decoded {
					result.Groups[name] = group
				}
			case "backdrops":
				decoded, err := d.DecodeNamesBackdrops(path, key, v)
				if err != nil {
					return result, err
				}
				for name, path := range decoded {
					result.Backdrops[name] = path
				}
			case "include":
				decoded, err := d.DecodeNamesIncludes(path, key, v)
				if err != nil {
					return result, err
				}
				for _, include := range decoded {
					for name, path := range include.Backdrops {
						result.Backdrops[name] = path
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

func (d *decoder) DecodeNamesBackdrops(path string, name string, config interface{}) (map[string]string, error) {
	result := map[string]string{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k := range t.Interface().(map[interface{}]interface{}) {
			key := k.(string)
			result[key] = path
		}
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}

func (d *decoder) DecodeNamesGroups(path string, name string, config interface{}) (map[string]Names, error) {
	result := map[string]Names{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			key := k.(string)
			decoded, err := d.DecodeNames(path, key, v)
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

func (d *decoder) DecodeNamesIncludes(path string, name string, config interface{}) ([]Names, error) {
	result := []Names{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		decoded, err := d.DecodeNamesInclude(path, name, config)
		if err != nil {
			return result, err
		}
		result = append(result, decoded)
	case reflect.Slice:
		for _, v := range t.Interface().([]interface{}) {
			decoded, err := d.DecodeNamesInclude(path, name, v)
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

func (d *decoder) DecodeNamesInclude(path string, name string, config interface{}) (Names, error) {
	var result Names
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "file":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				return d.includeFileNames(decoded)
			case "text":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				return d.includeTextNames(path, name, []byte(decoded))
			default:
				return result, &ConfigError{Name: name, UnsupportedKey: &key}
			}
		}
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}

func (d *decoder) includeFileNames(filename string) (Names, error) {
	if !filepath.IsAbs(filename) {
		directory, err := os.Getwd()
		if err != nil {
			return Names{}, err
		}
		filename, err = filepath.Abs(filepath.Join(directory, filename))
		if err != nil {
			return Names{}, err
		}
	}
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return Names{}, fmt.Errorf("could not read file '%s'", filename)
	}
	return d.WithFile(filename).includeTextNames(filename, filename, bytes)
}

func (d *decoder) includeTextNames(path string, name string, bytes []byte) (Names, error) {
	var mapType map[interface{}]interface{}
	err := yaml.Unmarshal(bytes, &mapType)
	if err != nil {
		return Names{}, err
	}
	config, err := d.DecodeNames(path, name, mapType)
	if err != nil {
		return Names{}, err
	}
	return config, nil
}
