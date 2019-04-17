package types

import (
	"fmt"
	"reflect"
	"strings"
)

// KeyValueList represents a list of key/value pairs
type KeyValueList []KeyValue

// KeyValue represents a key/value pair, where the value is optional
type KeyValue struct {
	Key   string
	Value *string
}

// Strings transforms a key/value list into a list of strings that will be
// understood by docker.
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

// DecodeKeyValueList creates key value configurations from a config map.
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
		return result, ErrorUnsupportedType(name, t.Kind())
	}
	return result, nil
}

// DecodeKeyValue creates a key value configuration from a config map.
func DecodeKeyValue(name string, config interface{}) (KeyValue, error) {
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		decoded, err := DecodeString(name, t.String())
		if err != nil {
			return KeyValue{}, err
		}
		switch values := strings.SplitN(decoded, "=", 2); len(values) {
		case 0:
			return KeyValue{}, fmt.Errorf("empty assignment in '%s'", name)
		case 1:
			return KeyValue{Key: values[0], Value: nil}, nil
		case 2:
			return KeyValue{Key: values[0], Value: &values[1]}, nil
		default:
			return KeyValue{}, fmt.Errorf("too many values in '%s'", name)
		}
	default:
		return KeyValue{}, ErrorUnsupportedType(name, t.Kind())
	}
}
