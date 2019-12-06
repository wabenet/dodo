package main

import (
	"github.com/hashicorp/go-plugin"
	"github.com/oclaussen/dodo/pkg/stages/dockermachine"
	"github.com/oclaussen/dodo/pkg/stages/grpc"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(new(log.JSONFormatter))
	plugin.Serve(&plugin.ServeConfig{
		GRPCServer:      plugin.DefaultGRPCServer,
		HandshakeConfig: grpc.HandshakeConfig("docker-machine"),
		Plugins: map[string]plugin.Plugin{
			"stage": &grpc.Plugin{Impl: &dockermachine.Stage{}},
		},
	})
}
