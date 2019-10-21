package designer

import (
	"encoding/json"
)

type Config struct {
	Hostname    string
	CA          string
	ServerCert  string
	ServerKey   string
	Environment []string
	DockerArgs  []string
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
