package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Filename string                   `yaml:"-"`
	Contexts map[string]ContextConfig `yaml:"contexts,omitempty"`
}

// TODO: validation
// TODO: check if there are unknown keys
func Load(filename string) (*Config, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Could not read file %q", filename)
	}

	config := Config{}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, fmt.Errorf("Could not load config from %q: %s", filename, err)
	}
	config.Filename = filename

	return &config, nil
}
