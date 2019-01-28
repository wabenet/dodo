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
	User          string
	Volumes       Volumes
	VolumesFrom   []string
	WorkingDir    string
	Interpreter   []string
	Script        string
	Command       []string
}

func decodeBackdrops(name string, config interface{}) (Backdrops, error) {
	result := map[string]BackdropConfig{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			key := k.(string)
			decoded, err := decodeBackdropConfig(key, v)
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

func decodeBackdropConfig(name string, config interface{}) (BackdropConfig, error) {
	var result BackdropConfig
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "build", "image":
				decoded, err := decodeBuild(key, v)
				if err != nil {
					return result, err
				}
				result.Build = &decoded
			case "container_name":
				decoded, err := decodeString(key, v)
				if err != nil {
					return result, err
				}
				result.ContainerName = decoded
			case "remove":
				decoded, err := decodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.Remove = &decoded
			case "interactive":
				decoded, err := decodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.Interactive = decoded
			case "environment":
				decoded, err := decodeKeyValueList(key, v)
				if err != nil {
					return result, err
				}
				result.Environment = decoded
			case "user":
				decoded, err := decodeString(key, v)
				if err != nil {
					return result, err
				}
				result.User = decoded
			case "volumes":
				decoded, err := decodeVolumes(key, v)
				if err != nil {
					return result, err
				}
				result.Volumes = decoded
			case "volumes_from":
				decoded, err := decodeStringSlice(key, v)
				if err != nil {
					return result, err
				}
				result.VolumesFrom = decoded
			case "working_dir":
				decoded, err := decodeString(key, v)
				if err != nil {
					return result, err
				}
				result.WorkingDir = decoded
			case "interpreter":
				decoded, err := decodeStringSlice(key, v)
				if err != nil {
					return result, err
				}
				result.Interpreter = decoded
			case "script":
				decoded, err := decodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Script = decoded
			case "command":
				decoded, err := decodeStringSlice(key, v)
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
