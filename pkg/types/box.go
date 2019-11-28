package types

import (
	"fmt"
	"reflect"
	"strings"
)

type Box struct {
	User        string
	Name        string
	Version     string
	AccessToken string
}

func (d *decoder) DecodeBox(name string, config interface{}) (Box, error) {
	var result Box
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		decoded, err := d.DecodeString(name, config)
		if err != nil {
			return result, err
		}
		switch values := strings.SplitN(decoded, "/", 2); len(values) {
		case 0:
			return result, fmt.Errorf("empty box name in '%s", name)
		case 1:
			result.Name = values[0]
		case 2:
			result.User = values[0]
			result.Name = values[1]
		default:
			return result, fmt.Errorf("invalid box name in '%s': %s", name, decoded)
		}
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "user", "username":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.User = decoded
			case "name":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Name = decoded
			case "version":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Version = decoded
			case "access_token":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.AccessToken = decoded
			default:
				return result, &ConfigError{Name: name, UnsupportedKey: &key}
			}
		}
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}
