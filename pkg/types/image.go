package types

import (
	"reflect"
)

type Image struct {
	Name         string
	Context      string
	Dockerfile   string
	Steps        []string
	Args         KeyValueList
	Secrets      Secrets
	SSHAgents    SSHAgents
	NoCache      bool
	ForceRebuild bool
	ForcePull    bool
	Requires     []string
}

func (target *Image) Merge(source *Image) {
	if len(source.Name) > 0 {
		target.Name = source.Name
	}
	if len(source.Context) > 0 {
		target.Context = source.Context
	}
	if len(source.Dockerfile) > 0 {
		target.Dockerfile = source.Dockerfile
	}
	if len(source.Steps) > 0 {
		target.Steps = source.Steps
	}
	target.Args = append(target.Args, source.Args...)
	target.Secrets = append(target.Secrets, source.Secrets...)
	target.SSHAgents = append(target.SSHAgents, source.SSHAgents...)
	if source.NoCache {
		target.NoCache = true
	}
	if source.ForceRebuild {
		target.ForceRebuild = true
	}
	if source.ForcePull {
		target.ForcePull = true
	}
	target.Requires = append(target.Requires, source.Requires...)
}

func (d *decoder) DecodeImage(name string, config interface{}) (Image, error) {
	var result Image
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		decoded, err := d.DecodeString(name, config)
		if err != nil {
			return result, err
		}
		result.Name = decoded
	case reflect.Map:
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "name":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Name = decoded
			case "context":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Context = decoded
			case "dockerfile":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Dockerfile = decoded
			case "steps", "inline":
				decoded, err := d.DecodeStringSlice(key, v)
				if err != nil {
					return result, err
				}
				result.Steps = decoded
			case "args", "arguments":
				decoded, err := d.DecodeKeyValueList(key, v)
				if err != nil {
					return result, err
				}
				result.Args = decoded
			case "secrets":
				decoded, err := d.DecodeSecrets(key, v)
				if err != nil {
					return result, err
				}
				result.Secrets = decoded
			case "ssh":
				decoded, err := d.DecodeSSHAgents(key, v)
				if err != nil {
					return result, err
				}
				result.SSHAgents = decoded
			case "no_cache":
				decoded, err := d.DecodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.NoCache = decoded
			case "force_rebuild":
				decoded, err := d.DecodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.ForceRebuild = decoded
			case "force_pull":
				decoded, err := d.DecodeBool(key, v)
				if err != nil {
					return result, err
				}
				result.ForcePull = decoded
			case "requires", "dependencies":
				decoded, err := d.DecodeStringSlice(key, v)
				if err != nil {
					return result, err
				}
				result.Requires = decoded
			default:
				return result, &ConfigError{Name: name, UnsupportedKey: &key}
			}
		}
	default:
		return result, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
	return result, nil
}
