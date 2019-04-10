package types

import (
	"reflect"
)

// Image represents the build configuration for a docker image
type Image struct {
	Name         string
	Context      string
	Dockerfile   string
	Steps        []string
	Args         KeyValueList
	NoCache      bool
	ForceRebuild bool
	ForcePull    bool
	PrintOutput  bool
}

func DecodeImage(name string, config interface{}) (Image, error) {
	var result Image
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		decoded, err := DecodeString(name, config)
		if err != nil {
			return result, err
		}
		result.Name = decoded
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "name":
				decoded, err := DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Name = decoded
			case "context":
				decoded, err := DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Context = decoded
			case "dockerfile":
				decoded, err := DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Dockerfile = decoded
			case "steps":
				decoded, err := DecodeStringSlice(key, v)
				if err != nil {
					return result, err
				}
				result.Steps = decoded
			case "args":
				decoded, err := DecodeKeyValueList(key, v)
				if err != nil {
					return result, err
				}
				result.Args = decoded
			case "no_cache":
				decoded, err := DecodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.NoCache = decoded
			case "force_rebuild":
				decoded, err := DecodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.ForceRebuild = decoded
			case "force_pull":
				decoded, err := DecodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.ForcePull = decoded
			default:
				return result, ErrorUnsupportedKey(name, key)
			}
		}
	default:
		return result, ErrorUnsupportedType(name, t.Kind())
	}
	return result, nil
}
