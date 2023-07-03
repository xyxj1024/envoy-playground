package snapshot

import (
	"context"
	"time"

	"envoy-swarm-control/pkg/logger"
	"envoy-swarm-control/pkg/xds"

	types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

type Manager struct {
	adsProvider   xds.ADS
	sdsProvider   xds.SDS
	snapshotCache cache.SnapshotCache
	logger        logger.Logger
}

func NewManager(ads xds.ADS, sds xds.SDS, config cache.SnapshotCache, log logger.Logger) *Manager {
	return &Manager{
		adsProvider:   ads,
		sdsProvider:   sds,
		snapshotCache: config,
		logger:        log,
	}
}

/* Function Listen:
 * just a wrapper around updateConfiguration.
 */
func (m *Manager) Listen(updateChannel chan UpdateReason) {
	for {
		update := <-updateChannel
		if err := m.updateConfiguration(update); err != nil {
			m.logger.Fatalf(err.Error()) // kill the application for now
		}

		time.Sleep(30 * time.Second)
	}
}

/* Function updateConfiguration:
 * fetches configuration resources from ADS and SDS providers every 30 seconds and
 * updates snapshot cache.
 */
func (m *Manager) updateConfiguration(update UpdateReason) error {
	m.logger.WithFields(logger.Fields{"reason": update}).
		Infof("Control plane updating configurations for Envoy")

	const discoveryInterval = 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), discoveryInterval)
	defer cancel()

	clusters, listeners, err := m.adsProvider.Provide(ctx, update.EnvoyListenerPort)
	if err != nil {
		return err
	}
	secrets, err := m.sdsProvider.Provide(ctx)
	if err != nil {
		return err
	}

	return m.generateSnapshot(clusters, listeners, secrets, update.EnvoyNodeId)
}

func (m *Manager) generateSnapshot(clusters, listeners, secrets []types.Resource, nodeId string) error {
	version := time.Now().Format(time.RFC3339) // timestamp as version number

	resources := make(map[string][]types.Resource, 3)
	resources[resource.ClusterType] = clusters
	resources[resource.ListenerType] = listeners
	resources[resource.SecretType] = secrets

	snap, err := cache.NewSnapshot(version, resources)
	if err != nil {
		return err
	}

	if err = snap.Consistent(); err != nil {
		return err
	}

	if err = m.snapshotCache.SetSnapshot(
		context.Background(),
		nodeId,
		snap,
	); err != nil {
		return err
	}
	m.logger.Infof("Snapshot served: %+v", snap)

	m.logger.WithFields(logger.Fields{
		"cluster-count":  len(clusters),
		"listener-count": len(listeners),
		"secrets-count":  len(secrets)}).Debugf("SnapshotCache updated")

	return err
}
