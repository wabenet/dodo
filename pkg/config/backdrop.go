package config

import (
	"reflect"
)

// Backdrops represents a mapping of backdrop names to backdrop configurations.
type Backdrops map[string]BackdropConfig

// BackdropConfig represents the configuration for a backdrop
// (possible target for running a command)
type BackdropConfig struct {
	Build         *BuildConfig
	ContainerName string
	Remove        *bool
	Pull          bool
	Interactive   bool
	Environment   KeyValueList
	Image         string
	User          string
	Volumes       []string
	VolumesFrom   []string
	WorkingDir    string
	Interpreter   []string
	Script        string
	Command       []string
}

func DecodeBackdrops(name string, config interface{}) (Backdrops, error) {
	result := map[string]BackdropConfig{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			key := k.(string)
			decoded, err := DecodeBackdropConfig(key, v)
			if err != nil {
				return result, err
			}
			result[key] = decoded
		}
	default:
		return result, errorUnsupportedType(name, t.Kind())
	}
	return result, nil
}

func DecodeBackdropConfig(name string, config interface{}) (BackdropConfig, error) {
	var result BackdropConfig
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "build":
				decoded, err := DecodeBuild(key, v)
				if err != nil {
					return result, err
				}
				result.Build = &decoded
			case "container_name":
				decoded, err := DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.ContainerName = decoded
			case "remove":
				decoded, err := DecodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.Remove = &decoded
			case "pull":
				decoded, err := DecodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.Pull = decoded
			case "interactive":
				decoded, err := DecodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.Interactive = decoded
			case "environment":
				decoded, err := DecodeKeyValueList(key, v)
				if err != nil {
					return result, err
				}
				result.Environment = decoded
			case "image":
				decoded, err := DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Image = decoded
			case "user":
				decoded, err := DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.User = decoded
			case "volumes":
				decoded, err := DecodeStringSlice(key, v)
				if err != nil {
					return result, err
				}
				result.Volumes = decoded
			case "volumes_from":
				decoded, err := DecodeStringSlice(key, v)
				if err != nil {
					return result, err
				}
				result.VolumesFrom = decoded
			case "working_dir":
				decoded, err := DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.WorkingDir = decoded
			case "interpreter":
				decoded, err := DecodeStringSlice(key, v)
				if err != nil {
					return result, err
				}
				result.Interpreter = decoded
			case "script":
				decoded, err := DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Script = decoded
			case "command":
				decoded, err := DecodeStringSlice(key, v)
				if err != nil {
					return result, err
				}
				result.Command = decoded
			default:
				return result, errorUnsupportedKey(name, key)
			}
		}
	default:
		return result, errorUnsupportedType(name, t.Kind())
	}
	return result, nil
}
