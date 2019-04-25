package types

import (
	"reflect"

	"github.com/oclaussen/dodo/pkg/template"
)

func DecodeBool(name string, config interface{}) (bool, error) {
	var result bool
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Bool:
		result = t.Bool()
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}

func DecodeString(name string, config interface{}) (string, error) {
	var result string
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		return template.ApplyTemplate(t.String())
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
}

func DecodeStringSlice(name string, config interface{}) ([]string, error) {
	var result []string
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		decoded, err := DecodeString(name, t.String())
		if err != nil {
			return result, err
		}
		result = []string{decoded}
	case reflect.Slice:
		for _, v := range t.Interface().([]interface{}) {
			decoded, err := DecodeString(name, v)
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
