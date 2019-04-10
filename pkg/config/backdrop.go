package config

import (
	"reflect"

	"github.com/oclaussen/dodo/pkg/image"
	"github.com/oclaussen/dodo/pkg/types"
)

// Backdrops represents a mapping of backdrop names to backdrop configurations.
type Backdrops map[string]BackdropConfig

// BackdropConfig represents the configuration for a backdrop
// (possible target for running a command)
type BackdropConfig struct {
	Image         *image.ImageConfig
	ContainerName string
	Remove        *bool
	Pull          bool
	Interactive   bool
	Environment   types.KeyValueList
	User          string
	Volumes       types.Volumes
	VolumesFrom   []string
	Ports         types.Ports
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
		return result, types.ErrorUnsupportedType(name, t.Kind())
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
				decoded, err := decodeImage(key, v)
				if err != nil {
					return result, err
				}
				result.Image = &decoded
			case "container_name":
				decoded, err := types.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.ContainerName = decoded
			case "remove":
				decoded, err := types.DecodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.Remove = &decoded
			case "interactive":
				decoded, err := types.DecodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.Interactive = decoded
			case "environment":
				decoded, err := types.DecodeKeyValueList(key, v)
				if err != nil {
					return result, err
				}
				result.Environment = decoded
			case "user":
				decoded, err := types.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.User = decoded
			case "volumes":
				decoded, err := types.DecodeVolumes(key, v)
				if err != nil {
					return result, err
				}
				result.Volumes = decoded
			case "volumes_from":
				decoded, err := types.DecodeStringSlice(key, v)
				if err != nil {
					return result, err
				}
				result.VolumesFrom = decoded
			case "ports":
				decoded, err := types.DecodePorts(key, v)
				if err != nil {
					return result, err
				}
				result.Ports = decoded
			case "working_dir":
				decoded, err := types.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.WorkingDir = decoded
			case "interpreter":
				decoded, err := types.DecodeStringSlice(key, v)
				if err != nil {
					return result, err
				}
				result.Interpreter = decoded
			case "script":
				decoded, err := types.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Script = decoded
			case "command":
				decoded, err := types.DecodeStringSlice(key, v)
				if err != nil {
					return result, err
				}
				result.Command = decoded
			default:
				return result, types.ErrorUnsupportedKey(name, key)
			}
		}
	default:
		return result, types.ErrorUnsupportedType(name, t.Kind())
	}
	return result, nil
}
