package types

import (
	"reflect"
)

type Stages map[string]Stage

type Stage struct {
	Type    string
	Options Options
}

type Options map[string]string

func DecodeStages(name string, config interface{}) (Stages, error) {
	result := map[string]Stage{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			key := k.(string)
			decoded, err := DecodeStage(key, v)
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

func DecodeStage(name string, config interface{}) (Stage, error) {
	var result Stage
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "type":
				decoded, err := DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Type = decoded
			case "options":
				decoded, err := DecodeOptions(key, v)
				if err != nil {
					return result, err
				}
				result.Options = decoded
			default:
				return result, &ConfigError{Name: name, UnsupportedKey: &key}
			}
		}
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}

func DecodeOptions(name string, config interface{}) (Options, error) {
	result := map[string]string{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			key := k.(string)
			decoded, err := DecodeString(key, v)
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
