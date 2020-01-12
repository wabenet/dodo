package types

import (
	"reflect"
)

type Resources struct {
	CPU     int64
	Memory  int64
	Volumes PersistentVolumes
	USB     USBFilters
}

type PersistentVolumes []PersistentVolume

type PersistentVolume struct {
	Size int64
}

type USBFilters []USBFilter

type USBFilter struct {
	Name      string
	VendorID  string
	ProductID string
}

func (d *decoder) DecodeResources(name string, config interface{}) (Resources, error) {
	result := Resources{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "cpu":
				decoded, err := d.DecodeInt(key, v)
				if err != nil {
					return result, err
				}
				result.CPU = decoded
			case "memory":
				decoded, err := d.DecodeBytes(key, v)
				if err != nil {
					return result, err
				}
				result.Memory = decoded
			case "volumes":
				decoded, err := d.DecodePersistentVolumes(name, v)
				if err != nil {
					return result, err
				}
				result.Volumes = decoded
			case "usb":
				decoded, err := d.DecodeUSBFilters(name, v)
				if err != nil {
					return result, err
				}
				result.USB = decoded
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

func (d *decoder) DecodeUSBFilters(name string, config interface{}) (USBFilters, error) {
	result := []USBFilter{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		decoded, err := d.DecodeUSBFilter(name, config)
		if err != nil {
			return result, err
		}
		result = append(result, decoded)
	case reflect.Slice:
		for _, v := range t.Interface().([]interface{}) {
			decoded, err := d.DecodeUSBFilter(name, v)
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

func (d *decoder) DecodeUSBFilter(name string, config interface{}) (USBFilter, error) {
	result := USBFilter{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "name":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Name = decoded
			case "vendorid":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.VendorID = decoded
			case "productid":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.ProductID = decoded
			default:
				return result, &ConfigError{Name: name, UnsupportedKey: &key}
			}
		}
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}
