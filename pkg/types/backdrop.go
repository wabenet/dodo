package types

import (
	"reflect"
)

type Backdrops map[string]Backdrop

type Backdrop struct {
	Name          string
	Aliases       []string
	Stage         string
	ForwardStage  bool
	Image         *Image
	ContainerName string
	Remove        *bool
	Pull          bool
	Interactive   bool
	Environment   KeyValueList
	User          string
	Volumes       Volumes
	VolumesFrom   []string
	Devices       Volumes // FIXME: this is a very lazy solution
	Ports         Ports
	WorkingDir    string
	Interpreter   []string
	Script        string
	Command       []string

	filename string
}

func (target *Backdrop) Merge(source *Backdrop) {
	if len(source.Name) > 0 {
		target.Name = source.Name
	}
	target.Aliases = append(target.Aliases, source.Aliases...)
	if len(source.Stage) > 0 {
		target.Stage = source.Stage
	}
	if source.ForwardStage {
		target.ForwardStage = true
	}
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
	target.Devices = append(target.Devices, source.Devices...)
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

func (d *decoder) DecodeBackdrops(name string, config interface{}) (Backdrops, error) {
	result := map[string]Backdrop{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			key := k.(string)
			decoded, err := d.DecodeBackdrop(key, v)
			if err != nil {
				return result, err
			}
			result[key] = decoded
			for _, alias := range decoded.Aliases {
				result[alias] = decoded
			}
		}
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}

func (d *decoder) DecodeBackdrop(name string, config interface{}) (Backdrop, error) {
	result := Backdrop{Name: name, filename: d.filename}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "alias", "aliases":
				decoded, err := d.DecodeStringSlice(key, v)
				if err != nil {
					return result, err
				}
				result.Aliases = decoded
			case "stage":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Stage = decoded
			case "forward_stage":
				decoded, err := d.DecodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.ForwardStage = decoded
			case "build", "image":
				decoded, err := d.DecodeImage(key, v)
				if err != nil {
					return result, err
				}
				result.Image = &decoded
			case "name", "container_name":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.ContainerName = decoded
			case "rm", "remove":
				decoded, err := d.DecodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.Remove = &decoded
			case "interactive":
				decoded, err := d.DecodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.Interactive = decoded
			case "env", "environment":
				decoded, err := d.DecodeKeyValueList(key, v)
				if err != nil {
					return result, err
				}
				result.Environment = decoded
			case "user":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.User = decoded
			case "volume", "volumes":
				decoded, err := d.DecodeVolumes(key, v)
				if err != nil {
					return result, err
				}
				result.Volumes = decoded
			case "volume_from", "volumes_from":
				decoded, err := d.DecodeStringSlice(key, v)
				if err != nil {
					return result, err
				}
				result.VolumesFrom = decoded
			case "device", "devices":
				decoded, err := d.DecodeVolumes(key, v)
				if err != nil {
					return result, err
				}
				result.Devices = decoded
			case "ports":
				decoded, err := d.DecodePorts(key, v)
				if err != nil {
					return result, err
				}
				result.Ports = decoded
			case "workdir", "working_dir":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.WorkingDir = decoded
			case "interpreter":
				decoded, err := d.DecodeStringSlice(key, v)
				if err != nil {
					return result, err
				}
				result.Interpreter = decoded
			case "script":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Script = decoded
			case "command", "arguments":
				decoded, err := d.DecodeStringSlice(key, v)
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
