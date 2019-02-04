package config

import (
	"reflect"

	"github.com/oclaussen/dodo/pkg/image"
)

func decodeImage(name string, config interface{}) (image.ImageConfig, error) {
	var result image.ImageConfig
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		decoded, err := decodeString(name, config)
		if err != nil {
			return result, err
		}
		result.Name = decoded
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "name":
				decoded, err := decodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Name = decoded
			case "context":
				decoded, err := decodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Context = decoded
			case "dockerfile":
				decoded, err := decodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Dockerfile = decoded
			case "steps":
				decoded, err := decodeStringSlice(key, v)
				if err != nil {
					return result, err
				}
				result.Steps = decoded
			case "args":
				decoded, err := decodeKeyValueList(key, v)
				if err != nil {
					return result, err
				}
				result.Args = decoded
			case "no_cache":
				decoded, err := decodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.NoCache = decoded
			case "force_rebuild":
				decoded, err := decodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.ForceRebuild = decoded
			case "force_pull":
				decoded, err := decodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.ForcePull = decoded
			default:
				return result, errorUnsupportedKey(name, key)
			}
		}
	default:
		return result, errorUnsupportedType(name, t.Kind())
	}
	return result, nil
}
