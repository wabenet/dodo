package config

import (
	"reflect"

	"github.com/oclaussen/dodo/pkg/types"
)

// Groups represents a mapping of group names to backdrop groups.
type Groups map[string]Group

// Group represents as set of backdrops, nested groups are allowed.
type Group struct {
	Backdrops types.Backdrops
	Groups    Groups
}

func decodeGroups(name string, config interface{}) (Groups, error) {
	result := map[string]Group{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			key := k.(string)
			decoded, err := decodeGroup(key, v)
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

func decodeGroup(name string, config interface{}) (Group, error) {
	result := Group{Backdrops: types.Backdrops{}, Groups: Groups{}}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "groups":
				decoded, err := decodeGroups(key, v)
				if err != nil {
					return result, err
				}
				for name, group := range decoded {
					result.Groups[name] = group
				}
			case "backdrops":
				decoded, err := types.DecodeBackdrops(key, v)
				if err != nil {
					return result, err
				}
				for name, backdrop := range decoded {
					result.Backdrops[name] = backdrop
				}
			case "include":
				decoded, err := decodeIncludes(key, v)
				if err != nil {
					return result, err
				}
				for _, include := range decoded {
					for name, backdrop := range include.Backdrops {
						result.Backdrops[name] = backdrop
					}
				}
			default:
				return result, types.ErrorUnsupportedKey(name, key)
			}
		}
	default:
		return result, types.ErrorUnsupportedType(name, t.Kind())
	}
	return result, nil
}
