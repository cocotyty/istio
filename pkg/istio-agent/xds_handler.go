package istioagent

import (
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
)

type XDSHandler interface {
	OnConnect(node *core.Node, send func(req *discovery.DiscoveryRequest))
	HandleResponse(resp *discovery.DiscoveryResponse) (stop bool, err error)
	OnDisconnect()
}
