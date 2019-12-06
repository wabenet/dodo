package main

import (
	"github.com/hashicorp/go-plugin"
	"github.com/oclaussen/dodo/pkg/stages/grpc"
	"github.com/oclaussen/dodo/pkg/stages/virtualbox"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(new(log.JSONFormatter))
	plugin.Serve(&plugin.ServeConfig{
		GRPCServer:      plugin.DefaultGRPCServer,
		HandshakeConfig: grpc.HandshakeConfig("virtualbox"),
		Plugins: map[string]plugin.Plugin{
			"stage": &grpc.Plugin{Impl: &virtualbox.Stage{}},
		},
	})
}
