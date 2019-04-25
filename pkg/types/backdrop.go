package types

import (
	"reflect"
)

type Backdrops map[string]Backdrop

type Backdrop struct {
	Image         *Image
	ContainerName string
	Remove        *bool
	Pull          bool
	Interactive   bool
	Environment   KeyValueList
	User          string
	Volumes       Volumes
	VolumesFrom   []string
	Ports         Ports
	WorkingDir    string
	Interpreter   []string
	Script        string
	Command       []string
}

func (target *Backdrop) Merge(source *Backdrop) {
	if source.Image != nil {
		target.Image.Merge(source.Image)
	}
	if len(source.ContainerName) > 0 {
		target.ContainerName = source.ContainerName
	}
	if source.Remove != nil {
		target.Remove = source.Remove
	}
	if source.Pull {
		target.Pull = true
	}
	if source.Interactive {
		target.Interactive = true
	}
	target.Environment = append(target.Environment, source.Environment...)
	if len(source.User) > 0 {
		target.User = source.User
	}
	target.Volumes = append(target.Volumes, source.Volumes...)
	target.VolumesFrom = append(target.VolumesFrom, source.VolumesFrom...)
	target.Ports = append(target.Ports, source.Ports...)
	if len(source.WorkingDir) > 0 {
		target.WorkingDir = source.WorkingDir
	}
	if len(source.Interpreter) > 0 {
		target.Interpreter = source.Interpreter
	}
	if len(source.Script) > 0 {
		target.Script = source.Script
	}
	if len(source.Command) > 0 {
		target.Command = source.Command
	}
}

func DecodeBackdrops(name string, config interface{}) (Backdrops, error) {
	result := map[string]Backdrop{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			key := k.(string)
			decoded, err := DecodeBackdrop(key, v)
			if err != nil {
				return result, err
			}
			result[key] = decoded
		}
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}

func DecodeBackdrop(name string, config interface{}) (Backdrop, error) {
	var result Backdrop
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "build", "image":
				decoded, err := DecodeImage(key, v)
				if err != nil {
					return result, err
				}
				result.Image = &decoded
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
			case "user":
				decoded, err := DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.User = decoded
			case "volumes":
				decoded, err := DecodeVolumes(key, v)
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
			case "ports":
				decoded, err := DecodePorts(key, v)
				if err != nil {
					return result, err
				}
				result.Ports = decoded
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
				return result, &ConfigError{Name: name, UnsupportedKey: &key}
			}
		}
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}
