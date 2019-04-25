package types

import (
	"reflect"

	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/sshforward/sshprovider"
)

type SSHAgents []SSHAgent

type SSHAgent struct {
	ID           string
	IdentityFile string
}

func (agents SSHAgents) SSHAgentProvider() (session.Attachable, error) {
	configs := make([]sshprovider.AgentConfig, 0, len(agents))
	for _, agent := range agents {
		config := sshprovider.AgentConfig{
			ID:    agent.ID,
			Paths: []string{agent.IdentityFile},
		}
		configs = append(configs, config)
	}
	return sshprovider.NewSSHAgentProvider(configs)
}

func DecodeSSHAgents(name string, config interface{}) (SSHAgents, error) {
	result := []SSHAgent{}
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.Bool:
		decoded, err := DecodeBool(name, config)
		if err != nil {
			return result, err
		}
		if decoded {
			result = append(result, SSHAgent{})
		}
	case reflect.String, reflect.Map:
		decoded, err := DecodeSSHAgent(name, config)
		if err != nil {
			return result, err
		}
		result = append(result, decoded)
	case reflect.Slice:
		for _, v := range t.Interface().([]interface{}) {
			decoded, err := DecodeSSHAgent(name, v)
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

func DecodeSSHAgent(name string, config interface{}) (SSHAgent, error) {
	switch t := reflect.ValueOf(config); t.Kind() {
	case reflect.String:
		decoded, err := DecodeKeyValue(name, config)
		if err != nil {
			return SSHAgent{}, err
		}
		result := SSHAgent{ID: decoded.Key}
		if decoded.Value != nil {
			result.IdentityFile = *decoded.Value
		}
		return result, nil
	case reflect.Map:
		var result SSHAgent
		for k, v := range t.Interface().(map[interface{}]interface{}) {
			switch key := k.(string); key {
			case "id":
				decoded, err := DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.ID = decoded
			case "file":
				decoded, err := DecodeString(key, v)
				if err != nil {
					return result, err
				}
				result.IdentityFile = decoded
			default:
				return result, &ConfigError{Name: name, UnsupportedKey: &key}
			}
		}
		return result, nil
	default:
		return SSHAgent{}, &ConfigError{Name: name, UnsupportedType: t.Kind()}
	}
}
