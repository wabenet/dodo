package main

import (
	"github.com/hashicorp/go-plugin"
	"github.com/oclaussen/dodo/pkg/stage"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(new(log.JSONFormatter))
	plugin.Serve(&plugin.ServeConfig{
		GRPCServer:      plugin.DefaultGRPCServer,
		HandshakeConfig: stage.HandshakeConfig("virtualbox"),
		Plugins: map[string]plugin.Plugin{
			"stage": &stage.Plugin{Impl: &Stage{}},
		},
	})
}
