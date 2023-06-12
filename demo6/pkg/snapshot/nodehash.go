package snapshot

import core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

type StaticHash struct{}

const staticHash = "test-id"

func (StaticHash) ID(node *core.Node) string {
	if node == nil {
		return ""
	}
	return staticHash
}
