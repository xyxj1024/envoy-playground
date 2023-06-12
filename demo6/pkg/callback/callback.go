package callback

import (
	// standard library
	"context"
	"sync"

	// Self-defined
	"envoy-demo4/pkg/logger"

	// Envoy go-control-plane
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	server "github.com/envoyproxy/go-control-plane/pkg/server/v3"
)

type Callbacks struct {
	Signal         chan struct{}
	Debug          bool
	Fetches        int
	Requests       int
	DeltaRequests  int
	DeltaResponses int
	mu             sync.Mutex // only one goroutine at a time can access callback structure
}

var (
	_ server.Callbacks = &Callbacks{}
	l logger.Logger
)

func (cb *Callbacks) Report() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	l.WithFields(logger.Fields{"fetches": cb.Fetches, "requests": cb.Requests})
}

func (cb *Callbacks) OnStreamOpen(_ context.Context, id int64, typ string) error {
	l.Infof("OnStreamOpen %d of type %s", id, typ)
	return nil
}

func (cb *Callbacks) OnStreamClosed(id int64, node *core.Node) {
	l.Infof("OnStreamClosed %d for node %s", id, node.Id)
}

func (cb *Callbacks) OnDeltaStreamOpen(_ context.Context, id int64, typ string) error {
	l.Infof("OnDeltaStreamOpen %d of type %s", id, typ)
	return nil
}

func (cb *Callbacks) OnDeltaStreamClosed(id int64, node *core.Node) {
	l.Infof("OnDeltaStreamClosed %d for node %s", id, node.Id)
}

func (cb *Callbacks) OnStreamRequest(id int64, req *discoverygrpc.DiscoveryRequest) error {
	l.Infof("OnStreamRequest %d Request [%v]", id, req.TypeUrl)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.Requests++
	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}
	return nil
}

func (cb *Callbacks) OnStreamResponse(ctx context.Context, id int64, req *discoverygrpc.DiscoveryRequest, res *discoverygrpc.DiscoveryResponse) {
	l.Infof("OnStreamResponse... %d Request [%v], Response [%v]", id, req.TypeUrl, res.TypeUrl)
	cb.Report()
}

func (cb *Callbacks) OnStreamDeltaResponse(id int64, req *discoverygrpc.DeltaDiscoveryRequest, res *discoverygrpc.DeltaDiscoveryResponse) {
	l.Infof("OnStreamDeltaResponse... %d Request [%v], Response [%v]", id, req.TypeUrl, res.TypeUrl)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.DeltaResponses++
}

func (cb *Callbacks) OnStreamDeltaRequest(id int64, req *discoverygrpc.DeltaDiscoveryRequest) error {
	l.Infof("OnStreamDeltaRequest... %d Request [%v]", id, req.TypeUrl)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.DeltaRequests++
	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}
	return nil
}

func (cb *Callbacks) OnFetchRequest(ctx context.Context, req *discoverygrpc.DiscoveryRequest) error {
	l.Infof("OnFetchRequest... Request [%v]", req.TypeUrl)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.Fetches++
	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}
	return nil
}

func (cb *Callbacks) OnFetchResponse(req *discoverygrpc.DiscoveryRequest, res *discoverygrpc.DiscoveryResponse) {
	l.Infof("OnFetchResponse... Request [%v], Response [%v]", req.TypeUrl, res.TypeUrl)
}
