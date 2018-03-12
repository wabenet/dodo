package config

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types/strslice"
)

// TODO: include other stuff (resources/networking)?
// TODO: inline dockerfile
// TODO: powerful entrypoint
// TODO: remove config
type CommandConfig struct {
	Build         *BuildConfig    `yaml:"build,omitempty"`
	ContainerName string          `yaml:"container_name,omitempty"`
	EnvFile       Stringorslice   `yaml:"env_file,omitempty"`
	Environment   MaporEqualSlice `yaml:"environment,omitempty"`
	Image         string          `yaml:"image,omitempty"`
	User          string          `yaml:"user,omitempty"`
	Volumes       *Volumes        `yaml:"volumes,omitempty"`
	VolumesFrom   []string        `yaml:"volumes_from,omitempty"`
	WorkingDir    string          `yaml:"working_dir,omitempty"`
}

type Stringorslice strslice.StrSlice

func (s *Stringorslice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var stringType string
	if err := unmarshal(&stringType); err == nil {
		*s = []string{stringType}
		return nil
	}

	var sliceType []interface{}
	if err := unmarshal(&sliceType); err == nil {
		parts, err := toStrings(sliceType)
		if err != nil {
			return err
		}
		*s = parts
		return nil
	}

	return errors.New("Failed to unmarshal Stringorslice")
}

type MaporEqualSlice []string

func (s *MaporEqualSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	parts, err := unmarshalToStringOrSepMapParts(unmarshal, "=")
	if err != nil {
		return err
	}
	*s = parts
	return nil
}

func unmarshalToStringOrSepMapParts(unmarshal func(interface{}) error, key string) ([]string, error) {
	var sliceType []interface{}
	if err := unmarshal(&sliceType); err == nil {
		return toStrings(sliceType)
	}
	var mapType map[interface{}]interface{}
	if err := unmarshal(&mapType); err == nil {
		return toSepMapParts(mapType, key)
	}
	return nil, errors.New("Failed to unmarshal MaporSlice")
}

func toSepMapParts(value map[interface{}]interface{}, sep string) ([]string, error) {
	if len(value) == 0 {
		return nil, nil
	}
	parts := make([]string, 0, len(value))
	for k, v := range value {
		if sk, ok := k.(string); ok {
			if sv, ok := v.(string); ok {
				parts = append(parts, sk+sep+sv)
			} else if sv, ok := v.(int); ok {
				parts = append(parts, sk+sep+strconv.Itoa(sv))
			} else if sv, ok := v.(int64); ok {
				parts = append(parts, sk+sep+strconv.FormatInt(sv, 10))
			} else if sv, ok := v.(float64); ok {
				parts = append(parts, sk+sep+strconv.FormatFloat(sv, 'f', -1, 64))
			} else if v == nil {
				parts = append(parts, sk)
			} else {
				return nil, fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", v, v)
			}
		} else {
			return nil, fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", k, k)
		}
	}
	return parts, nil
}

func toStrings(s []interface{}) ([]string, error) {
	if len(s) == 0 {
		return nil, nil
	}
	r := make([]string, len(s))
	for k, v := range s {
		if sv, ok := v.(string); ok {
			r[k] = sv
		} else {
			return nil, fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", v, v)
		}
	}
	return r, nil
}
