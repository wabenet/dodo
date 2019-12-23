package types

import (
	"reflect"
)

type Stages map[string]Stage

type Stage struct {
	Type    string
	Box     Box
	Volumes PersistentVolumes
	Options Options

	filename string
}

type PersistentVolumes []PersistentVolume

type PersistentVolume struct {
	Size int64
}

// TODO: this gives marshalling errors over grpc when used with nested maps
type Options map[string]interface{}

func (d *decoder) DecodeStages(name string, config interface{}) (Stages, error) {
	result := map[string]Stage{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			key := k.(string)
			decoded, err := d.DecodeStage(key, v)
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

func (d *decoder) DecodeStage(name string, config interface{}) (Stage, error) {
	result := Stage{filename: d.filename}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "type":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Type = decoded
			case "options":
				decoded, err := d.DecodeOptions(key, v)
				if err != nil {
					return result, err
				}
				result.Options = decoded
			case "box":
				decoded, err := d.DecodeBox(key, v)
				if err != nil {
					return result, err
				}
				result.Box = decoded
			case "volumes":
				decoded, err := d.DecodePersistentVolumes(key, v)
				if err != nil {
					return result, err
				}
				result.Volumes = decoded
			default:
				return result, &ConfigError{Name: name, UnsupportedKey: &key}
			}
		}
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}

func (d *decoder) DecodePersistentVolumes(name string, config interface{}) (PersistentVolumes, error) {
	result := []PersistentVolume{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		decoded, err := d.DecodePersistentVolume(name, config)
		if err != nil {
			return result, err
		}
		result = append(result, decoded)
	case reflect.Slice:
		for _, v := range t.Interface().([]interface{}) {
			decoded, err := d.DecodePersistentVolume(name, v)
			if err != nil {
				return result, err
			}
			result = append(result, decoded)
		}
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}

func (d *decoder) DecodePersistentVolume(name string, config interface{}) (PersistentVolume, error) {
	result := PersistentVolume{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "size":
				decoded, err := d.DecodeBytes(key, v)
				if err != nil {
					return result, err
				}
				result.Size = decoded
			}
		}
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}

func (d *decoder) DecodeOptions(name string, config interface{}) (Options, error) {
	result := map[string]interface{}{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			key := k.(string)
			result[key] = v
		}
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}
