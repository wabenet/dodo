package config

import (
	"reflect"
)

// Includes represents a list of items to include in the configuratio.
type Includes []Include

// Include represents something to include in the configuration, e.g. a path
// to a yaml file, or some plain yaml string.
type Include struct {
	File string
	Text string
}

func decodeIncludes(name string, config interface{}) (Includes, error) {
	result := []Include{}
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

func decodeInclude(name string, config interface{}) (Include, error) {
	var result Include
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "file":
				decoded, err := decodeString(key, v)
				if err != nil {
					return result, err
				}
				result.File = decoded
			case "text":
				decoded, err := decodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Text = decoded
			default:
				return result, errorUnsupportedKey(name, key)
			}
		}
	default:
		return result, errorUnsupportedType(name, t.Kind())
	}
	return result, nil
}
