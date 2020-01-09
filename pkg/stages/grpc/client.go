package grpc

import (
	"encoding/json"

	"github.com/oclaussen/dodo/pkg/stage"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/oclaussen/dodo/proto"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type GRPCClient struct {
	client proto.StageClient
}

func (client *GRPCClient) Initialize(name string, config *types.Stage) error {
	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "could not marshal json")
	}
	_, err = client.client.Initialize(context.Background(), &proto.InitRequest{Name: name, Config: string(jsonBytes)})
	return err
}

func (client *GRPCClient) Cleanup() {}

func (client *GRPCClient) Create() error {
	_, err := client.client.Create(context.Background(), &proto.Empty{})
	return err
}

func (client *GRPCClient) Remove(force bool, volumes bool) error {
	_, err := client.client.Remove(context.Background(), &proto.RemoveRequest{Force: force, Volumes: volumes})
	return err
}

func (client *GRPCClient) Start() error {
	_, err := client.client.Start(context.Background(), &proto.Empty{})
	return err
}

func (client *GRPCClient) Stop() error {
	_, err := client.client.Stop(context.Background(), &proto.Empty{})
	return err
}

func (client *GRPCClient) Exist() (bool, error) {
	response, err := client.client.Exist(context.Background(), &proto.Empty{})
	if err != nil {
		return false, err
	}
	return response.Exist, nil
}

func (client *GRPCClient) Available() (bool, error) {
	response, err := client.client.Available(context.Background(), &proto.Empty{})
	if err != nil {
		return false, err
	}
	return response.Available, nil
}

func (client *GRPCClient) GetSSHOptions() (*stage.SSHOptions, error) {
	response, err := client.client.GetSSHOptions(context.Background(), &proto.Empty{})
	if err != nil {
		return nil, err
	}
	return &stage.SSHOptions{
		Hostname:       response.Hostname,
		Port:           int(response.Port),
		Username:       response.Username,
		PrivateKeyFile: response.PrivateKeyFile,
	}, nil
}

func (client *GRPCClient) GetDockerOptions() (*stage.DockerOptions, error) {
	response, err := client.client.GetDockerOptions(context.Background(), &proto.Empty{})
	if err != nil {
		return nil, err
	}
	return &stage.DockerOptions{
		Version:  response.Version,
		Host:     response.Host,
		CAFile:   response.CaFile,
		CertFile: response.CertFile,
		KeyFile:  response.KeyFile,
	}, nil
}
