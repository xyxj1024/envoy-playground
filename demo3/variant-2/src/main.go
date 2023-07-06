package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"soloio-hoot-xds/xds"

	"go.uber.org/zap"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
)

type ClusterNodeHasher struct{}

// ID uses the node Cluster field
func (ClusterNodeHasher) ID(node *core.Node) string {
	if node == nil {
		return ""
	}
	return node.Cluster
}

func main() {
	ctx := context.Background()

	// Create a logger
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	l := logger.Sugar()

	// Use Envoy node's cluster name to set snapshot
	var nodeGroup string = "edge-gateway"

	// Create a cache
	cache := cachev3.NewSnapshotCache(
		false, // disable ADS
		ClusterNodeHasher{},
		l,
	)

	// Create the snapshot to be served by Envoy
	var version int32
	snapshot, err := xds.GenerateSnapshot(version, 0) // initial weight is zero
	if err != nil {
		l.Errorf("could not generate snapshot: %+v", err)
		os.Exit(1)
	}
	if err := snapshot.Consistent(); err != nil {
		l.Errorf("snapshot inconsistency: %+v\n%+v", snapshot, err)
		os.Exit(1)
	}
	l.Debugf("will serve snapshot %+v", snapshot)

	// Add the snapshot to the cache
	if err := cache.SetSnapshot(ctx, nodeGroup, snapshot); err != nil {
		l.Errorf("snapshot error %q for %+v", err, snapshot)
		os.Exit(1)
	}

	// Run the xDS server
	go RunServer(ctx, cache, xdsPort)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter weight for cluster-b: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		weight, _ := strconv.Atoi(text)
		clampedWeight := clampWeight(weight)
		fmt.Println("setting weight to", clampedWeight)

		snapshot, err = xds.GenerateSnapshot(version, clampedWeight)
		if err != nil {
			l.Errorf("could not generate snapshot: %+v", err)
			os.Exit(1)
		}
		if err := snapshot.Consistent(); err != nil {
			l.Errorf("snapshot inconsistency: %+v\n%+v", snapshot, err)
			os.Exit(1)
		}
		l.Debugf("will serve snapshot %+v", snapshot)

		// Add the snapshot to the cache
		if err := cache.SetSnapshot(ctx, nodeGroup, snapshot); err != nil {
			l.Errorf("snapshot error %q for %+v", err, snapshot)
			os.Exit(1)
		}
	}
}

func clampWeight(weight int) uint32 {
	// if weight > 100 {
	// 	weight = 100
	// }
	if weight < 0 {
		weight = 0
	}
	return uint32(weight)
}
