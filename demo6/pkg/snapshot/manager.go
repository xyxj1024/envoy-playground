package snapshot

import (
	// Standard library
	"context"
	"time"

	// Self-defined
	"envoy-demo4/pkg/logger"
	"envoy-demo4/pkg/provider"

	// Envoy go-control-plane
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

type Manager struct {
	adsProvider   provider.ADS
	sdsProvider   provider.SDS
	snapshotCache cache.SnapshotCache
	logger        logger.Logger
}

func NewManager(ads provider.ADS, sds provider.SDS, c cache.SnapshotCache, log logger.Logger) *Manager {
	return &Manager{
		adsProvider:   ads,
		sdsProvider:   sds,
		snapshotCache: c,
		logger:        log,
	}
}

func (m *Manager) Listen(updateChannel chan string) {
	for {
		updateReason := <-updateChannel
		if err := m.runServiceDiscovery(updateReason); err != nil {
			m.logger.Fatalf(err.Error()) // kill the application for now
		}
	}
}

func (m *Manager) runServiceDiscovery(updateReason string) error {
	m.logger.WithFields(logger.Fields{"reason": updateReason}).Infof("Running service discovery")

	const discoveryTimeout = 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), discoveryTimeout)
	defer cancel()

	clusters, listeners, err := m.adsProvider.Provide(ctx)
	if err != nil {
		return err
	}
	secrets, err := m.sdsProvider.Provide(ctx)
	if err != nil {
		return err
	}

	return m.generateSnapshot(clusters, listeners, secrets)
}

func (m *Manager) generateSnapshot(clusters, listeners, secrets []types.Resource) error {
	version := time.Now().Format(time.RFC3339) // timestamp as version number

	snap, err := cache.NewSnapshot(version, map[resource.Type][]types.Resource{
		resource.ClusterType:  clusters,
		resource.ListenerType: listeners,
		resource.SecretType:   secrets,
	})
	if err != nil {
		return err
	}

	if err = snap.Consistent(); err != nil {
		return err
	}

	if err = m.snapshotCache.SetSnapshot(
		context.Background(),
		staticHash, // the constant hash
		snap,
	); err != nil {
		return err
	}

	m.logger.WithFields(logger.Fields{"cluster-count": len(clusters), "listener-count": len(listeners), "secret-count": len(secrets)}).Debugf("Snapshot updated")

	return err
}
