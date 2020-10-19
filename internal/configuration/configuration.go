package configuration

import "github.com/NodeFactoryIo/vedran/pkg/http-tunnel/server"

type Configuration struct {
	AuthSecret          string
	Name                string
	Capacity            int64
	Whitelist           bool
	Fee                 float32
	Selection           string
	Port                int32
	PortPool            server.Pooler
	TunnelServerAddress string
}

var Config Configuration
