package designer

import (
	"encoding/base64"
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
	jsonBytes := make([]byte, base64.StdEncoding.DecodedLen(len(input)))
	if _, err := base64.StdEncoding.Decode(jsonBytes, input); err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(jsonBytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func EncodeConfig(config *Config) ([]byte, error) {
	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return []byte{}, err
	}
	b64Bytes := make([]byte, base64.StdEncoding.EncodedLen(len(jsonBytes)))
	base64.StdEncoding.Encode(b64Bytes, jsonBytes)
	return b64Bytes, nil
}
