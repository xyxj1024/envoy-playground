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
func (m *Manager) Listen(updateChannel chan string) {
	for {
		updateReason := <-updateChannel
		if err := m.updateConfiguration(updateReason); err != nil {
			m.logger.Fatalf(err.Error()) // kill the application for now
		}
	}
}

/* Function updateConfiguration:
 * fetches configuration resources from ADS and SDS providers every 30 seconds and
 * updates snapshot cache.
 */
func (m *Manager) updateConfiguration(updateReason string) error {
	m.logger.WithFields(logger.Fields{"reason": updateReason}).
		Infof("Control plane updating configurations for Envoy")

	const discoveryInterval = 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), discoveryInterval)
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
	var err error
	config := m.snapshotCache
	num := len(config.GetStatusKeys())
	if num > 0 {
		m.logger.Infof("%d connected nodes", num)
		for i := 0; i < num; i++ {
			nodeId := config.GetStatusKeys()[i]
			m.logger.Infof("creating snapshot for nodeID %s", nodeId)

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

			m.logger.WithFields(logger.Fields{
				"cluster-count":  len(clusters),
				"listener-count": len(listeners),
				"secret-count":   len(secrets),
			}).Debugf("SnapshotCache updated")
		}
	}

	return err
}
