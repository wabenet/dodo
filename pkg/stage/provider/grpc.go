package provider

import (
	"github.com/hashicorp/go-plugin"
	"github.com/oclaussen/dodo/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const ProtocolVersion = 1

var PluginMap = map[string]plugin.Plugin{
	"provider": &ProviderPlugin{},
}

func HandshakeConfig(providerName string) plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		ProtocolVersion:  ProtocolVersion,
		MagicCookieKey:   "DODO_STAGE_PROVIDER",
		MagicCookieValue: providerName,
	}
}

type ProviderPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl Provider
}

func (p *ProviderPlugin) GRPCServer(_ *plugin.GRPCBroker, server *grpc.Server) error {
	proto.RegisterProviderServer(server, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *ProviderPlugin) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, client *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewProviderClient(client)}, nil
}

type GRPCClient struct {
	client proto.ProviderClient
}

func (client *GRPCClient) Initialize(config map[string]string) (bool, error) {
	response, err := client.client.Initialize(context.Background(), &proto.InitRequest{Config: config})
	if err != nil {
		return false, err
	}
	return response.Success, nil
}

func (client *GRPCClient) Create() error {
	_, err := client.client.Create(context.Background(), &proto.Empty{})
	return err
}

func (client *GRPCClient) Remove() error {
	_, err := client.client.Remove(context.Background(), &proto.Empty{})
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

func (client *GRPCClient) GetIP() (string, error) {
	response, err := client.client.GetIP(context.Background(), &proto.Empty{})
	if err != nil {
		return "", err
	}
	return response.Ip, nil
}

func (client *GRPCClient) GetURL() (string, error) {
	response, err := client.client.GetURL(context.Background(), &proto.Empty{})
	if err != nil {
		return "", err
	}
	return response.Url, nil
}

func (client *GRPCClient) GetSSHOptions() (*SSHOptions, error) {
	response, err := client.client.GetSSHOptions(context.Background(), &proto.Empty{})
	if err != nil {
		return nil, err
	}
	return &SSHOptions{
		Hostname: response.Hostname,
		Port:     int(response.Port),
		Username: response.Username,
	}, nil
}

func (client *GRPCClient) GetDockerOptions() (*DockerOptions, error) {
	response, err := client.client.GetDockerOptions(context.Background(), &proto.Empty{})
	if err != nil {
		return nil, err
	}
	return &DockerOptions{
		Version:  response.Version,
		Host:     response.Host,
		CAFile:   response.CaFile,
		CertFile: response.CertFile,
		KeyFile:  response.KeyFile,
	}, nil
}

type GRPCServer struct {
	Impl Provider
}

func (server *GRPCServer) Initialize(ctx context.Context, request *proto.InitRequest) (*proto.InitResponse, error) {
	success, err := server.Impl.Initialize(request.Config)
	if err != nil {
		return nil, err
	}
	return &proto.InitResponse{Success: success}, nil
}

func (server *GRPCServer) Create(ctx context.Context, _ *proto.Empty) (*proto.Empty, error) {
	return &proto.Empty{}, server.Impl.Create()
}

func (server *GRPCServer) Remove(ctx context.Context, _ *proto.Empty) (*proto.Empty, error) {
	return &proto.Empty{}, server.Impl.Remove()
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

func (server *GRPCServer) GetIP(ctx context.Context, _ *proto.Empty) (*proto.IPResponse, error) {
	ip, err := server.Impl.GetIP()
	if err != nil {
		return nil, err
	}
	return &proto.IPResponse{Ip: ip}, nil
}

func (server *GRPCServer) GetURL(ctx context.Context, _ *proto.Empty) (*proto.URLResponse, error) {
	url, err := server.Impl.GetURL()
	if err != nil {
		return nil, err
	}
	return &proto.URLResponse{Url: url}, nil
}

func (server *GRPCServer) GetSSHOptions(ctx context.Context, _ *proto.Empty) (*proto.SSHOptionsResponse, error) {
	opts, err := server.Impl.GetSSHOptions()
	if err != nil {
		return nil, err
	}
	return &proto.SSHOptionsResponse{
		Hostname: opts.Hostname,
		Port:     int32(opts.Port),
		Username: opts.Username,
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
