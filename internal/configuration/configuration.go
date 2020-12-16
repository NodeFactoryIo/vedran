package configuration

import (
	"net/url"

	"github.com/NodeFactoryIo/vedran/pkg/http-tunnel/server"
)

type PayoutConfiguration struct {
	PayoutNumberOfDays int
	PayoutTotalReward  float64
	LbURL              *url.URL
}

type Configuration struct {
	AuthSecret          string
	Name                string
	CertFile            string
	KeyFile             string
	Capacity            int64
	WhitelistEnabled    bool
	Fee                 float32
	Selection           string
	Port                int32
	PortPool            server.Pooler
	TunnelServerAddress string
	PayoutConfiguration *PayoutConfiguration
}

var Config Configuration
