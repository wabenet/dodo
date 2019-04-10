package image

import (
	"testing"

	"github.com/moby/buildkit/client"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestBuildImage(t *testing.T) {
	displayCh := make(chan *client.SolveStatus)
	defer close(displayCh)
	image := fakeImage(t, &types.Image{
		Context: "./test",
	})
	result, err := image.runBuild(&contextData{
		remote:         "client-session",
		dockerfileName: "Dockerfile",
	}, displayCh)
	assert.Nil(t, err)
	assert.Equal(t, "NewImageID", result)
}

func TestBuildInlineImage(t *testing.T) {
	displayCh := make(chan *client.SolveStatus)
	defer close(displayCh)
	image := fakeImage(t, &types.Image{
		Steps: []string{"FROM scratch"},
	})
	result, err := image.runBuild(&contextData{
		remote:         "client-session",
		dockerfileName: "Dockerfile",
	}, displayCh)
	assert.Nil(t, err)
	assert.Equal(t, "NewImageID", result)
}
