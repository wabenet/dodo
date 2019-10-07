package types

import (
	"encoding/csv"
	"errors"
	"reflect"
	"strings"

	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/secrets/secretsprovider"
)

type Secrets []Secret

type Secret struct {
	ID   string
	Path string
}

func (secrets Secrets) SecretsProvider() (session.Attachable, error) {
	sources := make([]secretsprovider.FileSource, 0, len(secrets))
	for _, secret := range secrets {
		source := secretsprovider.FileSource{
			ID:       secret.ID,
			FilePath: secret.Path,
		}
		sources = append(sources, source)
	}
	store, err := secretsprovider.NewFileStore(sources)
	if err != nil {
		return nil, err
	}
	return secretsprovider.NewSecretProvider(store), nil
}

func (d *decoder) DecodeSecrets(name string, config interface{}) (Secrets, error) {
	result := []Secret{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String, reflect.Map:
		decoded, err := d.DecodeSecret(name, config)
		if err != nil {
			return result, err
		}
		result = append(result, decoded)
	case reflect.Slice:
		for _, v := range t.Interface().([]interface{}) {
			decoded, err := d.DecodeSecret(name, v)
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

func (d *decoder) DecodeSecret(name string, config interface{}) (Secret, error) {
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		decoded, err := d.DecodeString(name, t.String())
		if err != nil {
			return Secret{}, err
		}

		reader := csv.NewReader(strings.NewReader(decoded))
		fields, err := reader.Read()
		if err != nil {
			return Secret{}, err
		}

		secretMap := make(map[interface{}]interface{}, len(fields))
		for _, field := range fields {
			kv, err := d.DecodeKeyValue(name, field)
			if err != nil {
				return Secret{}, err
			}
			if kv.Value == nil {
				return Secret{}, errors.New("invalid format for secrets")
			}
			secretMap[kv.Key] = *kv.Value
		}
		return d.DecodeSecret(name, secretMap)
	case reflect.Map:
		var result Secret
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "id":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.ID = decoded
			case "source", "src":
				decoded, err := d.DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.Path = decoded
			default:
				return result, &ConfigError{Name: name, UnsupportedKey: &key}
			}
		}
		return result, nil
	default:
		return Secret{}, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
}
