package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

const simplePull = `
image: testimage
`

func TestSimplePull(t *testing.T) {
	config := getExampleConfig(t, simplePull)
	assert.NotNil(t, config.Image)
	assert.Equal(t, "testimage", config.Image.Name)
}

const contextOnly = `
image:
  context: ./path/to/context
`

func TestContextOnly(t *testing.T) {
	config := getExampleConfig(t, contextOnly)
	assert.NotNil(t, config.Image)
	assert.Equal(t, "./path/to/context", config.Image.Context)
}

const environments = `
environment:
  - FOO=BAR
  - SOMETHING
`

func TestEnvironments(t *testing.T) {
	config := getExampleConfig(t, environments)
	assert.Equal(t, 2, len(config.Environment))

	foobar := config.Environment[0]
	assert.Equal(t, foobar.Key, "FOO")
	assert.NotNil(t, foobar.Value)
	assert.Equal(t, "BAR", *foobar.Value)

	something := config.Environment[1]
	assert.Equal(t, "SOMETHING", something.Key)
	assert.Nil(t, something.Value)
}

const simpleVolume = `
volumes: foo:bar:ro
`

func TestSimpleVolume(t *testing.T) {
	config := getExampleConfig(t, simpleVolume)
	assert.Equal(t, 1, len(config.Volumes))
	assert.Equal(t, "foo", config.Volumes[0].Source)
	assert.Equal(t, "bar", config.Volumes[0].Target)
	assert.True(t, config.Volumes[0].ReadOnly)
}

const mixedVolumes = `
volumes:
  - test
  - source: from
    target: to
    read_only: true
  - bar:baz
`

func TestMixedVolumes(t *testing.T) {
	config := getExampleConfig(t, mixedVolumes)
	assert.Equal(t, 3, len(config.Volumes))

	sourceOnly := config.Volumes[0]
	assert.Equal(t, "test", sourceOnly.Source)
	assert.Equal(t, "", sourceOnly.Target)
	assert.False(t, sourceOnly.ReadOnly)

	fullSpec := config.Volumes[1]
	assert.Equal(t, "from", fullSpec.Source)
	assert.Equal(t, "to", fullSpec.Target)
	assert.True(t, fullSpec.ReadOnly)

	readWrite := config.Volumes[2]
	assert.Equal(t, "bar", readWrite.Source)
	assert.Equal(t, "baz", readWrite.Target)
	assert.False(t, readWrite.ReadOnly)
}

const fullExample = `
image:
  name: testimage
  context: .
  dockerfile: Dockerfile
  steps:
    - RUN hello
    - RUN world
  args:
    - FOO=BAR
  no_cache: true
  force_rebuild: true
  force_pull: true
container_name: testcontainer
remove: false
interactive: true
volumes_from: 'somevolume'
interpreter: '/bin/sh'
script: |
  echo "$@"
command: ['Hello', 'World']
`

func TestFullExample(t *testing.T) {
	config := getExampleConfig(t, fullExample)
	assert.NotNil(t, config.Image)
	assert.Equal(t, "testimage", config.Image.Name)
	assert.Equal(t, ".", config.Image.Context)
	assert.Equal(t, "Dockerfile", config.Image.Dockerfile)
	assert.Equal(t, []string{"RUN hello", "RUN world"}, config.Image.Steps)
	assert.Equal(t, 1, len(config.Image.Args))
	assert.Equal(t, "FOO", config.Image.Args[0].Key)
	assert.Equal(t, "BAR", *config.Image.Args[0].Value)
	assert.True(t, config.Image.NoCache)
	assert.True(t, config.Image.ForceRebuild)
	assert.True(t, config.Image.ForcePull)
	assert.Equal(t, "testcontainer", config.ContainerName)
	assert.NotNil(t, config.Remove)
	assert.False(t, *config.Remove)
	assert.True(t, config.Interactive)
	assert.Contains(t, config.VolumesFrom, "somevolume")
	assert.Contains(t, config.Interpreter, "/bin/sh")
	assert.Equal(t, "echo \"$@\"\n", config.Script)
	assert.Equal(t, []string{"Hello", "World"}, config.Command)
}

func getExampleConfig(t *testing.T, yamlConfig string) Backdrop {
	var mapType map[interface{}]interface{}
	err := yaml.Unmarshal([]byte(yamlConfig), &mapType)
	assert.Nil(t, err)
	decoder := NewDecoder("example", "")
	config, err := decoder.DecodeBackdrop("example", mapType)
	assert.Nil(t, err)
	return config
}
