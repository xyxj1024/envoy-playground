package snapshot

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	myresource "envoy-swarm-control/pkg/resource"

	"github.com/sirupsen/logrus"

	types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

type Manager struct {
	snapshotCache               cache.SnapshotCache
	clusters, listeners, routes []types.Resource
}

func NewManager(config cache.SnapshotCache) *Manager {
	return &Manager{
		snapshotCache: config,
		clusters:      []types.Resource{},
		listeners:     []types.Resource{},
		routes:        []types.Resource{},
	}
}

/* Function Discover:
 * just a wrapper around updateConfiguration.
 */
func (m *Manager) Discover(updateChannel chan ServiceLabels, ctx context.Context) {
	for {
		update := <-updateChannel
		if reflect.DeepEqual(update, ServiceLabels{}) {
			continue
		}

		m.updateConfiguration(update, ctx)

		time.Sleep(30 * time.Second)
	}
}

func (m *Manager) updateConfiguration(update ServiceLabels, ctx context.Context) {
	version := time.Now().Format(time.RFC3339) // timestamp as version number
	logrus.Infof(">>>>>>>>>>>>>>>>>>> creating snapshot " + fmt.Sprint(version) + " for nodeID " + fmt.Sprint(update.Status.NodeID))

	cluster := myresource.ProvideCluster(
		fmt.Sprintf("%s_cluster", update.Status.NodeID),
		update.Route.UpstreamHost,
		update.Endpoint.Port.PortValue,
	)
	listener := myresource.ProvideHTTPListener(
		fmt.Sprintf("%s_listener", update.Status.NodeID),
		fmt.Sprintf("%s_route", update.Status.NodeID),
		update.Listener.Port.PortValue,
	)
	route := myresource.ProvideRoute(
		fmt.Sprintf("%s_route", update.Status.NodeID),
		fmt.Sprintf("%s_service", update.Status.NodeID),
		fmt.Sprintf("%s_cluster", update.Status.NodeID),
		update.Route.UpstreamHost,
	)

	m.clusters = append(m.clusters, cluster)
	m.listeners = append(m.listeners, listener)
	m.routes = append(m.routes, route)

	resources := make(map[string][]types.Resource, 3)
	resources[resource.ClusterType] = []types.Resource{cluster}
	resources[resource.RouteType] = []types.Resource{route}
	resources[resource.ListenerType] = []types.Resource{listener}

	snap, _ := cache.NewSnapshot(fmt.Sprint(version), resources)
	if err := snap.Consistent(); err != nil {
		logrus.Errorf("Snapshot inconsistency: %+v\n%+v", snap, err)
		os.Exit(1)
	}

	if err := m.snapshotCache.SetSnapshot(ctx, update.Status.NodeID, snap); err != nil {
		logrus.Fatalf("Snapshot error %q for %+v", err, snap)
	}

	logrus.Infof("Snapshot served: %+v", snap)
}
