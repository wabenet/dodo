package grpc

import (
	"encoding/json"

	"github.com/hashicorp/go-plugin"
	"github.com/oclaussen/dodo/pkg/stage"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/oclaussen/dodo/proto"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const ProtocolVersion = 1

var PluginMap = map[string]plugin.Plugin{
	"stage": &Plugin{},
}

func HandshakeConfig(stageType string) plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		ProtocolVersion:  ProtocolVersion,
		MagicCookieKey:   "DODO_STAGE",
		MagicCookieValue: stageType,
	}
}

type Plugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl stage.Stage
}

func (p *Plugin) GRPCServer(_ *plugin.GRPCBroker, server *grpc.Server) error {
	proto.RegisterStageServer(server, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *Plugin) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, client *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewStageClient(client)}, nil
}

type GRPCClient struct {
	client proto.StageClient
}

func (client *GRPCClient) Initialize(name string, config *types.Stage) (bool, error) {
	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return false, errors.Wrap(err, "could not marshal json")
	}
	response, err := client.client.Initialize(context.Background(), &proto.InitRequest{Name: name, Config: string(jsonBytes)})
	if err != nil {
		return false, err
	}
	return response.Success, nil
}

func (client *GRPCClient) Create() error {
	_, err := client.client.Create(context.Background(), &proto.Empty{})
	return err
}

func (client *GRPCClient) Remove(force bool) error {
	_, err := client.client.Remove(context.Background(), &proto.RemoveRequest{Force: force})
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

type GRPCServer struct {
	Impl stage.Stage
}

func (server *GRPCServer) Initialize(ctx context.Context, request *proto.InitRequest) (*proto.InitResponse, error) {
	var config types.Stage
	if err := json.Unmarshal([]byte(request.Config), &config); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal json")
	}
	success, err := server.Impl.Initialize(request.Name, &config)
	if err != nil {
		return nil, err
	}
	return &proto.InitResponse{Success: success}, nil
}

func (server *GRPCServer) Create(ctx context.Context, _ *proto.Empty) (*proto.Empty, error) {
	return &proto.Empty{}, server.Impl.Create()
}

func (server *GRPCServer) Remove(ctx context.Context, request *proto.RemoveRequest) (*proto.Empty, error) {
	return &proto.Empty{}, server.Impl.Remove(request.Force)
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
