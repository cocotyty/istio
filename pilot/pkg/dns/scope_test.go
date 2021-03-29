package dns

import (
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"istio.io/pkg/log"
	"testing"
	"time"
)

func TestLocalEgressScope_Update(t *testing.T) {
	log.FindScope(log.DefaultScopeName).SetOutputLevel(log.DebugLevel)
	scope := NewLocalEgressScope()
	scope.OnConnect(nil, func(req *discovery.DiscoveryRequest) {
		log.Debug(req)
	})
	go func() {
		time.Sleep(time.Millisecond * 100)
		scope.handleResponse([]string{"reviews.default.svc.cluster.local."})
	}()
	err := scope.Update("reviews.default.svc.cluster.local.")
	if err != nil {
		t.Fail()
	}
	err = scope.Update("reviews.default.svc.cluster.local.")
	if err != nil {
		t.Fail()
	}

	err = scope.Update("reviews.default.svc.cluster.local.")
	if err != nil {
		t.Fail()
	}
}
