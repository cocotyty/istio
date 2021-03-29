package dns

import (
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"log"
	"testing"
	"time"
)

func TestLocalEgressScope_Update(t *testing.T) {
	scope := NewLocalEgressScope()
	scope.OnConnect(nil, func(req *discovery.DiscoveryRequest) {
		log.Println(req)
	})
	go func() {
		time.Sleep(time.Millisecond * 100)
		scope.handleResponse([]string{"reviews.default.svc.cluster.local."})
	}()
	err := scope.Update("reviews.default.svc.cluster.local.")
	if err != nil {
		t.Fail()
	}
	go func() {
		time.Sleep(time.Millisecond * 100)
		scope.handleResponse([]string{"reviews.default.svc.cluster.local."})
	}()
	err = scope.Update("reviews.default.svc.cluster.local.")
	if err != nil {
		t.Fail()
	}
	go func() {
		time.Sleep(time.Millisecond * 100)
		scope.handleResponse([]string{"reviews.default.svc.cluster.local."})
	}()
	err = scope.Update("reviews.default.svc.cluster.local.")
	if err != nil {
		t.Fail()
	}
}
