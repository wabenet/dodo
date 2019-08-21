package designer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testConfig = &Config{
	Hostname:    "example",
	CA:          "someca",
	ServerCert:  "someCert",
	ServerKey:   "someKey",
	Environment: []string{"FOO=bar"},
}

func TestConfig(t *testing.T) {
	bytes, err := EncodeConfig(testConfig)
	assert.Nil(t, err)
	result, err := DecodeConfig(bytes)
	assert.Nil(t, err)
	assert.Equal(t, result, testConfig)
}
