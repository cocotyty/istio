package store

import (
	networking "istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/dns/schemas"
	"strings"
	"testing"
)

func TestEgressStore(t *testing.T) {
	AppendToEgressScope(map[string]string{
		"app": "reviews",
	}, "default", "ratings.default.svc.cluster.local.")
	AppendToEgressScope(map[string]string{
		"app": "reviews",
	}, "default", "pages.default.svc.cluster.local.")
	cfg, _ := EgressStore().List(schemas.EgressSidecarGVK, "")
	for i, host := range cfg[0].Spec.(*networking.Sidecar).Egress[0].Hosts {
		host = strings.TrimPrefix(host, "*/")
		host += "."
		switch i {
		case 0:
			if host != "ratings.default.svc.cluster.local." {
				t.Fail()
			}
		case 1:
			if host != "pages.default.svc.cluster.local." {
				t.Fail()
			}
		}
	}
}
