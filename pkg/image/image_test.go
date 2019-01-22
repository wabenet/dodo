package image

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestLocalImage(t *testing.T) {
	client := fakeImageClient{
		t:            t,
		existLocally: true,
		expectList:   true,
		expectPull:   false,
		expectBuild:  false,
	}
	options := Options{
		Name:   "localimage",
		Client: &client,
	}
	result, err := Get(context.Background(), options)
	assert.Nil(t, err)
	assert.Equal(t, "localimage", result)
	client.checkAssertions()
}

func TestRemoteImage(t *testing.T) {
	client := fakeImageClient{
		t:               t,
		existRemotelyAs: "docker.io/library/remoteimage:latest",
		expectList:      true,
		expectPull:      true,
		expectBuild:     false,
	}
	options := Options{
		Name:   "remoteimage",
		Client: &client,
	}
	result, err := Get(context.Background(), options)
	assert.Nil(t, err)
	assert.Equal(t, "docker.io/library/remoteimage:latest", result)
	client.checkAssertions()
}

func TestForcePull(t *testing.T) {
	client := fakeImageClient{
		t:               t,
		existLocally:    true,
		existRemotelyAs: "docker.io/library/remoteimage:latest",
		expectList:      false,
		expectPull:      true,
		expectBuild:     false,
	}
	options := Options{
		Name:      "remoteimage",
		ForcePull: true,
		Client:    &client,
	}
	result, err := Get(context.Background(), options)
	assert.Nil(t, err)
	assert.Equal(t, "docker.io/library/remoteimage:latest", result)
	client.checkAssertions()
}

func TestMissingImage(t *testing.T) {
	client := fakeImageClient{
		t:           t,
		expectList:  true,
		expectPull:  true,
		expectBuild: false,
	}
	options := Options{
		Name:   "missing",
		Client: &client,
	}
	result, err := Get(context.Background(), options)
	assert.Equal(t, errors.New("image does not exist"), err)
	assert.Equal(t, "", result)
	client.checkAssertions()
}

func TestBuildImage(t *testing.T) {
	client := fakeImageClient{
		t:           t,
		willBuildAs: "NewImageID",
		expectList:  false,
		expectPull:  false,
		expectBuild: true,
	}
	options := Options{
		DoBuild: true,
		Context: "./test",
		Client:  &client,
	}
	result, err := Get(context.Background(), options)
	assert.Nil(t, err)
	assert.Equal(t, "NewImageID", result)
	client.checkAssertions()
}

func TestBuildInlineImage(t *testing.T) {
	client := fakeImageClient{
		t:           t,
		willBuildAs: "NewImageID",
		expectList:  false,
		expectPull:  false,
		expectBuild: true,
	}
	options := Options{
		DoBuild: true,
		Steps:   []string{"FROM scratch"},
		Client:  &client,
	}
	result, err := Get(context.Background(), options)
	assert.Nil(t, err)
	assert.Equal(t, "NewImageID", result)
	client.checkAssertions()
}

func TestUnspecifiedImage(t *testing.T) {
	client := fakeImageClient{
		t:           t,
		expectList:  false,
		expectPull:  false,
		expectBuild: false,
	}
	options := Options{
		Client: &client,
	}
	result, err := Get(context.Background(), options)
	expected := errors.New("you need to specify either image name or build")
	assert.Equal(t, expected, err)
	assert.Equal(t, "", result)
	client.checkAssertions()
}

type fakeImageClient struct {
	t               *testing.T
	existLocally    bool
	existRemotelyAs string
	willBuildAs     string
	expectList      bool
	expectPull      bool
	expectBuild     bool
	didList         bool
	didPull         bool
	didBuild        bool
}

func (client *fakeImageClient) checkAssertions() {
	assert.Equal(client.t, client.expectList, client.didList)
	assert.Equal(client.t, client.expectPull, client.didPull)
	assert.Equal(client.t, client.expectBuild, client.didBuild)
}

func (client *fakeImageClient) Info(_ context.Context) (types.Info, error) {
	return types.Info{}, nil
}

func (client *fakeImageClient) ClientVersion() string {
	return "1.39"
}

func (client *fakeImageClient) DialSession(
	_ context.Context, _ string, _ map[string][]string,
) (net.Conn, error) {
	return nil, nil
}

func (client *fakeImageClient) BuildCancel(_ context.Context, _ string) error {
	return nil
}

func (client *fakeImageClient) ImageList(
	_ context.Context, _ types.ImageListOptions,
) ([]types.ImageSummary, error) {
	assert.Equal(client.t, true, client.expectList)
	assert.Equal(client.t, false, client.didList)
	client.didList = true
	if client.existLocally {
		return []types.ImageSummary{types.ImageSummary{}}, nil
	}
	return []types.ImageSummary{}, nil
}

func (client *fakeImageClient) ImagePull(
	_ context.Context, _ string, _ types.ImagePullOptions,
) (io.ReadCloser, error) {
	assert.Equal(client.t, true, client.expectPull)
	assert.Equal(client.t, false, client.didPull)
	client.didPull = true
	if client.existRemotelyAs == "" {
		return nil, errors.New("image does not exist")
	}
	message := jsonmessage.JSONMessage{
		Stream: "hello world",
	}
	response, err := json.Marshal(message)
	assert.Nil(client.t, err)
	return ioutil.NopCloser(bytes.NewReader(response)), nil
}

func (client *fakeImageClient) ImageBuild(
	_ context.Context, _ io.Reader, _ types.ImageBuildOptions,
) (types.ImageBuildResponse, error) {
	assert.Equal(client.t, true, client.expectBuild)
	assert.Equal(client.t, false, client.didBuild)
	client.didBuild = true
	buildResult := types.BuildResult{ID: client.willBuildAs}
	auxJSON, err := json.Marshal(buildResult)
	assert.Nil(client.t, err)
	rawJSON := json.RawMessage(auxJSON)
	message := jsonmessage.JSONMessage{
		ID:     "moby.image.id",
		Stream: "hello world",
		Aux:    &rawJSON,
	}
	response, err := json.Marshal(message)
	assert.Nil(client.t, err)
	return types.ImageBuildResponse{Body: ioutil.NopCloser(bytes.NewReader(response))}, nil
}
