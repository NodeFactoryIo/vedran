package script

import "net/url"

var stats, _ = url.Parse("/api/v1/stats")
var ws, _ = url.Parse("/ws")

func statsEndpoint(loadbalancerUrl *url.URL) *url.URL {
	return loadbalancerUrl.ResolveReference(stats)
}

func wsEndpoint(loadbalancerUrl *url.URL) *url.URL {
	loadbalancerWsUrl := loadbalancerUrl.ResolveReference(ws)
	loadbalancerWsUrl.Scheme = "ws"
	return loadbalancerWsUrl
}
