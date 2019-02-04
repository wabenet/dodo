package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/oclaussen/dodo/pkg/types"
)

func decodeKeyValueList(name string, config interface{}) (types.KeyValueList, error) {
	result := []types.KeyValue{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		decoded, err := decodeKeyValue(name, config)
		if err != nil {
			return result, err
		}
		result = append(result, decoded)
	case reflect.Slice:
		for _, v := range t.Interface().([]interface{}) {
			decoded, err := decodeKeyValueList(name, v)
			if err != nil {
				return result, err
			}
			result = append(result, decoded...)
		}
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			key := k.(string)
			decoded, err := decodeString(key, v)
			if err != nil {
				return result, err
			}
			result = append(result, types.KeyValue{
				Key:   key,
				Value: &decoded,
			})
		}
	default:
		return result, errorUnsupportedType(name, t.Kind())
	}
	return result, nil
}

func decodeKeyValue(name string, config interface{}) (types.KeyValue, error) {
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		decoded, err := decodeString(name, t.String())
		if err != nil {
			return types.KeyValue{}, err
		}
		switch values := strings.SplitN(decoded, "=", 2); len(values) {
		case 0:
			return types.KeyValue{}, fmt.Errorf("Empty assignment in '%s'", name)
		case 1:
			return types.KeyValue{Key: values[0], Value: nil}, nil
		case 2:
			return types.KeyValue{Key: values[0], Value: &values[1]}, nil
		default:
			return types.KeyValue{}, fmt.Errorf("Too many values in '%s'", name)
		}
	default:
		return types.KeyValue{}, errorUnsupportedType(name, t.Kind())
	}
}
