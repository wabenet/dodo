package designer

import (
	"encoding/json"
)

type Config struct {
	Hostname          string
	Environment       []string
	DockerArgs        []string
	DefaultUser       string
	AuthorizedSSHKeys []string
	Script            []string
}

type ProvisionResult struct {
	IPAddress  string
	CA         string
	ClientCert string
	ClientKey  string
}

func DecodeConfig(input []byte) (*Config, error) {
	var config Config
	if err := json.Unmarshal(input, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func EncodeConfig(config *Config) ([]byte, error) {
	return json.Marshal(config)
}

func DecodeResult(input []byte) (*ProvisionResult, error) {
	var result ProvisionResult
	if err := json.Unmarshal(input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func EncodeResult(result *ProvisionResult) ([]byte, error) {
	return json.Marshal(result)
}
