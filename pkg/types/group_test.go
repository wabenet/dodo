package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

const groupedYaml = `
groups:
  first:
    backdrops:
      example1:
        image: testimage

    groups:
      second:
        backdrops:
          example2:
            image: testimage

  third:
    backdrops:
      example3:
        image: testimage
`

func TestNestedGroups(t *testing.T) {
	var mapType map[interface{}]interface{}
	err := yaml.Unmarshal([]byte(groupedYaml), &mapType)
	assert.Nil(t, err)
	decoder := NewDecoder("example", "")
	config, err := decoder.DecodeGroup("example", mapType)
	assert.Nil(t, err)
	assert.Contains(t, config.Groups, "first")
	assert.Contains(t, config.Groups["first"].Backdrops, "example1")
	assert.Contains(t, config.Groups["first"].Groups, "second")
	assert.Contains(t, config.Groups["first"].Groups["second"].Backdrops, "example2")
	assert.Contains(t, config.Groups, "third")
	assert.Contains(t, config.Groups["third"].Backdrops, "example3")
}
