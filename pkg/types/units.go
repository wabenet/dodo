package types

import (
	"reflect"

	"github.com/alecthomas/units"
)

func (d *decoder) DecodeBytes(name string, config interface{}) (int64, error) {
	var result int64
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		v, err := d.ApplyTemplate(t.String())
		if err != nil {
			return result, err
		}
		result, err = units.ParseStrictBytes(v)
		if err != nil {
			return result, err
		}
		return result, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		result = t.Int()
		return result, nil
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}
