package config

import (
	"fmt"
	"reflect"
	"strings"
)

type KeyValueList []KeyValue

type KeyValue struct {
	Key   string
	Value *string
}

func (kvs *KeyValueList) Strings() []string {
	result := []string{}
	for _, kv := range *kvs {
		result = append(result, kv.String())
	}
	return result
}

func (kv *KeyValue) String() string {
	if kv.Value == nil {
		return kv.Key
	}
	return fmt.Sprintf("%s=%s", kv.Key, *kv.Value)
}

func DecodeKeyValueList(name string, config interface{}) (KeyValueList, error) {
	result := []KeyValue{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		decoded, err := DecodeKeyValue(name, config)
		if err != nil {
			return result, err
		}
		result = append(result, decoded)
	case reflect.Slice:
		for _, v := range t.Interface().([]interface{}) {
			decoded, err := DecodeKeyValueList(name, v)
			if err != nil {
				return result, err
			}
			result = append(result, decoded...)
		}
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			key := k.(string)
			decoded, err := DecodeString(key, v)
			if err != nil {
				return result, err
			}
			result = append(result, KeyValue{
				Key:   key,
				Value: &decoded,
			})
		}
	default:
		return result, errorUnsupportedType(name, t.Kind())
	}
	return result, nil
}

func DecodeKeyValue(name string, config interface{}) (KeyValue, error) {
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		switch values := strings.SplitN(t.String(), "=", 2); len(values) {
		case 0:
			return KeyValue{}, fmt.Errorf("Empty assignment in '%s'", name)
		case 1:
			return KeyValue{Key: values[0], Value: nil}, nil
		case 2:
			return KeyValue{Key: values[0], Value: &values[1]}, nil
		default:
			return KeyValue{}, fmt.Errorf("Too many values in '%s'", name)
		}
	default:
		return KeyValue{}, errorUnsupportedType(name, t.Kind())
	}
}
