package grpc

import (
	"encoding/json"

	"github.com/oclaussen/dodo/pkg/stage"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/oclaussen/dodo/proto"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type GRPCServer struct {
	Impl stage.Stage
}

func (server *GRPCServer) Initialize(ctx context.Context, request *proto.InitRequest) (*proto.Empty, error) {
	var config types.Stage
	if err := json.Unmarshal([]byte(request.Config), &config); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal json")
	}
	return &proto.Empty{}, server.Impl.Initialize(request.Name, &config)
}

func (server *GRPCServer) Create(ctx context.Context, _ *proto.Empty) (*proto.Empty, error) {
	return &proto.Empty{}, server.Impl.Create()
}

func (server *GRPCServer) Remove(ctx context.Context, request *proto.RemoveRequest) (*proto.Empty, error) {
	return &proto.Empty{}, server.Impl.Remove(request.Force, request.Volumes)
}

func (server *GRPCServer) Start(ctx context.Context, _ *proto.Empty) (*proto.Empty, error) {
	return &proto.Empty{}, server.Impl.Start()
}

func (server *GRPCServer) Stop(ctx context.Context, _ *proto.Empty) (*proto.Empty, error) {
	return &proto.Empty{}, server.Impl.Stop()
}

func (server *GRPCServer) Exist(ctx context.Context, _ *proto.Empty) (*proto.ExistResponse, error) {
	exist, err := server.Impl.Exist()
	if err != nil {
		return nil, err
	}
	return &proto.ExistResponse{Exist: exist}, nil
}

func (server *GRPCServer) Available(ctx context.Context, _ *proto.Empty) (*proto.AvailableResponse, error) {
	available, err := server.Impl.Available()
	if err != nil {
		return nil, err
	}
	return &proto.AvailableResponse{Available: available}, nil
}

func (server *GRPCServer) GetSSHOptions(ctx context.Context, _ *proto.Empty) (*proto.SSHOptionsResponse, error) {
	opts, err := server.Impl.GetSSHOptions()
	if err != nil {
		return nil, err
	}
	return &proto.SSHOptionsResponse{
		Hostname:       opts.Hostname,
		Port:           int32(opts.Port),
		Username:       opts.Username,
		PrivateKeyFile: opts.PrivateKeyFile,
	}, nil
}

func (server *GRPCServer) GetDockerOptions(ctx context.Context, _ *proto.Empty) (*proto.DockerOptionsResponse, error) {
	opts, err := server.Impl.GetDockerOptions()
	if err != nil {
		return nil, err
	}
	return &proto.DockerOptionsResponse{
		Version:  opts.Version,
		Host:     opts.Host,
		CaFile:   opts.CAFile,
		CertFile: opts.CertFile,
		KeyFile:  opts.KeyFile,
	}, nil
}
