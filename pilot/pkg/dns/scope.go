package dns

import (
	"regexp"
	"strings"
	"sync"
	"time"
)

func NewLocalEgressScope(sharedScopeName string) *LocalEgressScope {
	return &LocalEgressScope{
		scope:           map[string]bool{},
		mu:              sync.RWMutex{},
		sharedScopeName: sharedScopeName,
		req:             make(chan *EgressScopeRequest, 1),
		waiters:         map[string][]chan<- struct{}{},
	}
}

type LocalEgressScope struct {
	scope           map[string]bool
	mu              sync.RWMutex
	sharedScopeName string
	req             chan *EgressScopeRequest
	waiters         map[string][]chan<- struct{}
}

var hostNameReg = regexp.MustCompile(``)

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
		return nil
	}
	l.mu.Lock()
	if l.scope[host] {
		l.mu.Unlock()
		return nil
	}
	ch := make(chan struct{}, 1)
	waiters := append(l.waiters[host], ch)
	l.waiters[host] = waiters
	if len(waiters) == 1 {
		l.req <- &EgressScopeRequest{
			ScopeName: l.sharedScopeName,
			Hosts:     []string{host},
		}
	}
	l.mu.Unlock()
	timeout := time.NewTimer(time.Second * 2)
	select {
	case <-timeout.C:
		l.mu.Lock()
		waiters := l.waiters[host]
		for i, c := range waiters {
			if c != ch {
				continue
			}
			waiters = append(waiters[:i], waiters[i+1:]...)
			if len(waiters) == 0 {
				delete(l.waiters, host)
			} else {
				l.waiters[host] = waiters
			}
		}
		l.mu.Unlock()
		return nil
	case <-ch:
		if !timeout.Stop() {
			<-timeout.C
		}
		return nil
	}
}

type EgressScopeRequest struct {
	ScopeName string
	Hosts     []string
}

type EgressScopeResponse EgressScopeRequest

func (l *LocalEgressScope) Request() (req <-chan *EgressScopeRequest) {
	return l.req
}

func (l *LocalEgressScope) Response(resp *EgressScopeResponse) {
	m := make(map[string]bool, len(resp.Hosts))
	for _, host := range resp.Hosts {
		m[host] = true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.scope = m
	for host := range m {
		waiters := l.waiters[host]
		if len(waiters) == 0 {
			continue
		}
		delete(l.waiters, host)
		for _, w := range waiters {
			w <- struct{}{}
		}
	}
}
