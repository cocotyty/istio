// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xds

import (
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"istio.io/istio/pilot/pkg/dns/store"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/networking/util"
	istio_networking_nds_v1 "istio.io/istio/pilot/pkg/proto"
	"istio.io/pkg/log"
	"strings"
	"sync"
)

// Nds stands for EgressScope Discovery Service.
type EsdsGenerator struct {
	mu sync.Mutex
}

func (e *EsdsGenerator) Handle(req *discovery.DiscoveryRequest, proxy *model.Proxy) error {
	if proxy.Type != model.SidecarProxy {
		return nil
	}
	names := proxy.Metadata.ProxyConfig.ProxyMetadata["ISTIO_META_SHARED_DNS_EGRESS_SCOPE"]
	if names == "" {
		log.Debug("ISTIO_META_SHARED_DNS_EGRESS_SCOPE is empty")
		return nil
	}
	nameList := strings.Split(names, ",")
	labels := make(map[string]string, len(nameList))
	for _, name := range nameList {
		labels[name] = proxy.Metadata.Labels[name]
	}
	store.AppendToEgressScope(labels, proxy.ConfigNamespace, req.ResourceNames...)
	return nil
}

func (e *EsdsGenerator) Generate(proxy *model.Proxy, push *model.PushContext, w *model.WatchedResource, updates *model.PushRequest) model.Resources {
	if proxy.Type != model.SidecarProxy {
		return nil
	}
	var hosts []string
	if proxy.DNSEgressSidecarScope != nil && len(proxy.DNSEgressSidecarScope.EgressListeners) > 0 {
		hosts = proxy.DNSEgressSidecarScope.EgressListeners[0].IstioListener.Hosts
	}
	return model.Resources{
		util.MessageToAny(
			&istio_networking_nds_v1.EgressScope{Hosts: hosts},
		),
	}
}

var _ model.XdsResourceGenerator = &EsdsGenerator{}
