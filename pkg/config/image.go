package config

import (
	"reflect"

	"github.com/oclaussen/dodo/pkg/image"
	"github.com/oclaussen/dodo/pkg/types"
)

func decodeImage(name string, config interface{}) (image.ImageConfig, error) {
	var result image.ImageConfig
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		decoded, err := types.DecodeString(name, config)
		if err != nil {
			return result, err
		}
		result.Name = decoded
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "name":
				decoded, err := types.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Name = decoded
			case "context":
				decoded, err := types.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Context = decoded
			case "dockerfile":
				decoded, err := types.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Dockerfile = decoded
			case "steps":
				decoded, err := types.DecodeStringSlice(key, v)
				if err != nil {
					return result, err
				}
				result.Steps = decoded
			case "args":
				decoded, err := types.DecodeKeyValueList(key, v)
				if err != nil {
					return result, err
				}
				result.Args = decoded
			case "no_cache":
				decoded, err := types.DecodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.NoCache = decoded
			case "force_rebuild":
				decoded, err := types.DecodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.ForceRebuild = decoded
			case "force_pull":
				decoded, err := types.DecodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.ForcePull = decoded
			default:
				return result, types.ErrorUnsupportedKey(name, key)
			}
		}
	default:
		return result, types.ErrorUnsupportedType(name, t.Kind())
	}
	return result, nil
}
