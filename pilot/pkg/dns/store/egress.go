package store

import (
	networking "istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/config/memory"
	"istio.io/istio/pilot/pkg/dns/schemas"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/config"
	"istio.io/pkg/log"
	"sort"
	"strings"
	"sync"
)

var egressStoreCache = memory.NewController(memory.MakeSkipValidation(schemas.EgressScopeSchemas, false))
var egressStoreLocker = sync.Mutex{}

func EgressStore() model.ConfigStoreCache {
	return egressStoreCache
}

func AppendToEgressScope(labels map[string]string, namespace string, hosts ...string) {
	log.Infof("append %v to egress scope [%s/%v]", hosts, labels, namespace)
	for i, host := range hosts {
		host = "*/" + strings.TrimRight(host, ".")
		hosts[i] = host
	}
	egressStoreLocker.Lock()
	defer egressStoreLocker.Unlock()
	nameList := make([]string, 0, len(labels))
	for name := range labels {
		nameList = append(nameList, name)
	}
	sort.Strings(nameList)
	cfgName := ""
	for _, name := range nameList {
		cfgName += name + ":" + labels[name]
	}
	c := egressStoreCache.Get(schemas.EgressSidecarGVK, cfgName, namespace)
	if c == nil {
		egressStoreCache.Create(config.Config{
			Meta: config.Meta{
				GroupVersionKind: schemas.EgressSidecarGVK,
				Name:             cfgName,
				Namespace:        namespace,
			},
			Spec: &networking.Sidecar{
				WorkloadSelector: &networking.WorkloadSelector{Labels: labels},
				Egress: []*networking.IstioEgressListener{
					{Hosts: hosts},
				},
			},
		})
	} else {
		egress := c.Spec.(*networking.Sidecar).Egress[0]
		egress.Hosts = appendHosts(egress.Hosts, hosts)
		egressStoreCache.Update(*c)
	}
}

func appendHosts(hosts []string, more []string) []string {
	m := make(map[string]struct{}, len(hosts)+len(more))
	for _, host := range hosts {
		m[host] = struct{}{}
	}
	for _, host := range more {
		_, ok := m[host]
		if !ok {
			hosts = append(hosts, host)
			m[host] = struct{}{}
		}
	}
	return hosts
}
