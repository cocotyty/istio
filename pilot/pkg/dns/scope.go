package dns

import (
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"golang.org/x/sync/singleflight"
	istio_networking_nds_v1 "istio.io/istio/pilot/pkg/proto"
	v3 "istio.io/istio/pilot/pkg/xds/v3"
	"istio.io/pkg/log"
	"strings"
	"sync"
	"time"
)

func NewLocalEgressScope() *LocalEgressScope {
	le := &LocalEgressScope{
		scope:    map[string]bool{},
		mu:       sync.RWMutex{},
		waitSync: map[string]bool{},
		queue:    make(chan chan struct{}, 100),
	}
	go le.loopSend()
	return le
}

type LocalEgressScope struct {
	queue    chan chan struct{}
	scope    map[string]bool
	waitSync map[string]bool
	mu       sync.RWMutex
	nonce    string
	node     *core.Node
	send     func(req *discovery.DiscoveryRequest)
	group    singleflight.Group
	ws       []chan struct{}
}

func (l *LocalEgressScope) OnConnect(node *core.Node, send func(req *discovery.DiscoveryRequest)) {
	l.mu.Lock()
	l.node = node
	l.send = send
	l.mu.Unlock()
	l.sendRequest()
}

func (l *LocalEgressScope) sendRequest() {
	l.mu.Lock()
	hosts := make([]string, 0, len(l.scope)+len(l.waitSync))
	for host := range l.scope {
		hosts = append(hosts, host)
	}
	for host := range l.waitSync {
		hosts = append(hosts, host)
	}
	l.waitSync = map[string]bool{}
	send := l.send
	l.mu.Unlock()
	log.Debugf("local egress scope send request: %v", hosts)
	req := &discovery.DiscoveryRequest{
		ResourceNames: hosts,
		TypeUrl:       v3.EgressScopeType,
		ResponseNonce: l.nonce,
	}
	if send != nil {
		l.send(req)
	}
}

func (l *LocalEgressScope) HandleResponse(resp *discovery.DiscoveryResponse) (stop bool, err error) {
	for _, resource := range resp.Resources {
		scope := &istio_networking_nds_v1.EgressScope{}
		_ = resource.UnmarshalTo(scope)
		l.handleResponse(scope.Hosts)
	}
	return true, nil
}

func (l *LocalEgressScope) OnDisconnect() {
}

func isClusterService(host string) bool {
	parts := strings.SplitN(host, ".", 3)
	if len(parts) == 3 && parts[2] == "svc.cluster.local." {
		return true
	}
	return false
}

func (l *LocalEgressScope) Update(host string) error {
	if !isClusterService(host) {
		return nil
	}

	l.mu.RLock()
	exist := l.scope[host]
	l.mu.RUnlock()
	if exist {
		log.Debug("host already exist:", host)
		return nil
	}
	l.mu.Lock()
	l.waitSync[host] = true
	l.mu.Unlock()

	ch := make(chan struct{}, 1)
	timeout := time.NewTimer(time.Second * 2)
	log.Debug("add host to egress scope:", host)
	l.sendRequest()
	l.mu.Lock()
	l.ws = append(l.ws, ch)
	l.mu.Unlock()

	select {
	case <-ch:
		if !timeout.Stop() {
			<-timeout.C
		}
	case <-timeout.C:
		log.Debug("wait timeout")
	}
	log.Debug("add host to egress scope done:", host)
	return nil
}

const debounceTime = time.Millisecond * 20
const maxDebounceTime = time.Millisecond * 100

func (l *LocalEgressScope) loopSend() error {
	log.Debug("start send loop")
	maxTimer := time.NewTimer(maxDebounceTime)
	timer := time.NewTimer(debounceTime)
	var ws []chan struct{}
	send := func() {
		if len(ws) == 0 {
			return
		}
		l.sendRequest()
		l.mu.Lock()
		l.ws = append(l.ws, ws...)
		l.mu.Unlock()
		ws = nil
	}
	defer log.Debug("stop send loop")
	for {
		select {
		case w := <-l.queue:
			ws = append(ws, w)
			log.Debug("new hosts request")
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(debounceTime)
		case <-timer.C:
			send()
			if !maxTimer.Stop() {
				<-maxTimer.C
			}
			maxTimer.Reset(maxDebounceTime)
		case <-maxTimer.C:
			send()
			maxTimer.Reset(maxDebounceTime)
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(debounceTime)
		}
	}
}

func (l *LocalEgressScope) handleResponse(hosts []string) {
	log.Debugf("handle response: %v", hosts)
	m := make(map[string]bool, len(hosts))
	for _, host := range hosts {
		m[host] = true
	}
	l.mu.Lock()
	l.scope = m
	ws := l.ws
	l.ws = nil
	l.mu.Unlock()

	for _, w := range ws {
		w <- struct{}{}
	}
}
