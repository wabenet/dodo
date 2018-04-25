package config

import (
	"reflect"
)

func DecodeBool(name string, config interface{}) (bool, error) {
	var result bool
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Bool:
		result = t.Bool()
	default:
		return result, errorUnsupportedType(name, t.Kind())
	}
	return result, nil
}

func DecodeString(name string, config interface{}) (string, error) {
	var result string
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		result = t.String()
	default:
		return result, errorUnsupportedType(name, t.Kind())
	}
	return result, nil
}

func DecodeStringSlice(name string, config interface{}) ([]string, error) {
	var result []string
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		result = []string{t.String()}
	case reflect.Slice:
		for _, v := range t.Interface().([]interface{}) {
			decoded, err := DecodeString(name, v)
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
