package store

import (
	"istio.io/istio/pilot/pkg/dns/schemas"
	"testing"
)

func TestEgressStore(t *testing.T) {
	AppendToEgressScope(map[string]string{
		"app": "reviews",
	}, "default", "ratings.default.svc.cluster.local.")
	AppendToEgressScope(map[string]string{
		"app": "reviews",
	}, "default", "pages.default.svc.cluster.local.")
	t.Log(EgressStore().List(schemas.EgressSidecarGVK, ""))

}
