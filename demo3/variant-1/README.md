## Development Environment

```bash
$ sw_vers
ProductName:	macOS
ProductVersion:	12.6.5
BuildVersion:	21G531

$ go version
go version go1.20.4 darwin/amd64

$ envoy --version | grep -o '[0-9].[0-9]\+.[0-9]' | tail -n1
1.25.1
```


## Run Demo Code

- Generate TLS certificates and keys:

    ```bash
    $ cd certs && ./certs.sh
    ```

- Build Go module:

    ```bash
    $ cd src
    $ go mod init envoy-demo3 && go mod tidy
    $ go build
    ```

- View the results:

    ```bash
    # Terminal 1: Start control plane
    $ cd src
    $ GRPC_GO_LOG_VERBOSITY_LEVEL=99 GRPC_GO_LOG_SEVERITY_LEVEL=info go run envoy-demo3

    # Terminal 2: Start Envoy proxy
    $ envoy -c config/dynamic_config.yaml -l debug

    # Terminal 3: Access proxy
    $ curl -H "Host: http.domain.com" \
        --resolve http.domain.com:10000:127.0.0.1 \
        --cacert certs/envoy-intermediate-ca.crt \
        https://http.domain.com:10000/
    ```

## Envoy Debugging Information

After Envoy is started, we can verify our initial configuration (`dynamic_config.yaml`) through the following console output:

```bash
[2023-05-22 18:18:41.967][1042742][info][main] [source/server/server.cc:819] runtime: {}
[2023-05-22 18:18:41.971][1042742][info][admin] [source/server/admin/admin.cc:67] admin address: 127.0.0.1:9000
[2023-05-22 18:18:41.971][1042742][info][config] [source/server/configuration_impl.cc:131] loading tracing configuration
[2023-05-22 18:18:41.971][1042742][info][config] [source/server/configuration_impl.cc:91] loading 0 static secret(s)
[2023-05-22 18:18:41.971][1042742][info][config] [source/server/configuration_impl.cc:97] loading 1 cluster(s)
[2023-05-22 18:18:41.972][1042748][debug][grpc] [source/common/grpc/google_async_client_impl.cc:51] completionThread running
[2023-05-22 18:18:41.978][1042742][debug][config] [./source/common/http/filter_chain_helper.h:88]     upstream http filter #0
[2023-05-22 18:18:41.978][1042742][debug][config] [./source/common/http/filter_chain_helper.h:118]       name: envoy.filters.http.upstream_codec
[2023-05-22 18:18:41.978][1042742][debug][config] [./source/common/http/filter_chain_helper.h:121]     config: {"@type":"type.googleapis.com/envoy.extensions.filters.http.upstream_codec.v3.UpstreamCodec"}
[2023-05-22 18:18:41.980][1042742][debug][upstream] [source/common/upstream/upstream_impl.cc:451] transport socket match, socket default selected for host with address 127.0.0.1:18000
[2023-05-22 18:18:41.983][1042742][debug][upstream] [source/common/upstream/upstream_impl.cc:1446] initializing Primary cluster xds_cluster completed
[2023-05-22 18:18:41.983][1042742][debug][init] [source/common/init/manager_impl.cc:49] init manager Cluster xds_cluster contains no targets
[2023-05-22 18:18:41.983][1042742][debug][init] [source/common/init/watcher_impl.cc:14] init manager Cluster xds_cluster initialized, notifying ClusterImplBase
[2023-05-22 18:18:41.983][1042742][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:1117] adding TLS cluster xds_cluster
[2023-05-22 18:18:41.986][1042742][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:1195] membership update for TLS cluster xds_cluster added 1 removed 0
[2023-05-22 18:18:41.986][1042742][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:149] cm init: init complete: cluster=xds_cluster primary=0 secondary=0
[2023-05-22 18:18:41.986][1042742][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:169] maybe finish initialize state: 0
[2023-05-22 18:18:41.986][1042742][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:115] cm init: adding: cluster=xds_cluster primary=0 secondary=0
[2023-05-22 18:18:41.986][1042742][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:169] maybe finish initialize state: 1
[2023-05-22 18:18:41.986][1042742][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:178] maybe finish initialize primary init clusters empty: true
[2023-05-22 18:18:41.988][1042742][debug][config] [./source/common/config/grpc_stream.h:62] Establishing new gRPC bidi stream to xds_cluster for rpc StreamAggregatedResources(stream .envoy.service.discovery.v3.DiscoveryRequest) returns (stream .envoy.service.discovery.v3.DiscoveryResponse);

[2023-05-22 18:18:41.989][1042742][debug][router] [source/common/router/router.cc:470] [C0][S13139667687758040534] cluster 'xds_cluster' match for URL '/envoy.service.discovery.v3.AggregatedDiscoveryService/StreamAggregatedResources'
[2023-05-22 18:18:41.989][1042742][debug][router] [source/common/router/router.cc:678] [C0][S13139667687758040534] router decoding headers:
':method', 'POST'
':path', '/envoy.service.discovery.v3.AggregatedDiscoveryService/StreamAggregatedResources'
':authority', 'xds_cluster'
':scheme', 'http'
'te', 'trailers'
'content-type', 'application/grpc'
'x-envoy-internal', 'true'
'x-forwarded-for', '192.168.1.248'

[2023-05-22 18:18:41.990][1042742][debug][pool] [source/common/http/conn_pool_base.cc:78] queueing stream due to no available connections (ready=0 busy=0 connecting=0)
[2023-05-22 18:18:41.990][1042742][debug][pool] [source/common/conn_pool/conn_pool_base.cc:291] trying to create new connection
[2023-05-22 18:18:41.990][1042742][debug][pool] [source/common/conn_pool/conn_pool_base.cc:145] creating a new connection (connecting=0)
[2023-05-22 18:18:41.991][1042742][debug][http2] [source/common/http/http2/codec_impl.cc:1597] [C0] updating connection-level initial window size to 268435456
[2023-05-22 18:18:41.991][1042742][debug][connection] [./source/common/network/connection_impl.h:92] [C0] current connecting state: true
[2023-05-22 18:18:41.991][1042742][debug][client] [source/common/http/codec_client.cc:57] [C0] connecting
[2023-05-22 18:18:41.991][1042742][debug][connection] [source/common/network/connection_impl.cc:939] [C0] connecting to 127.0.0.1:18000
[2023-05-22 18:18:41.991][1042742][debug][connection] [source/common/network/connection_impl.cc:958] [C0] connection in progress
[2023-05-22 18:18:41.992][1042742][info][config] [source/server/configuration_impl.cc:101] loading 0 listener(s)
[2023-05-22 18:18:41.992][1042742][info][config] [source/server/configuration_impl.cc:113] loading stats configuration
[2023-05-22 18:18:41.992][1042742][debug][init] [source/common/init/manager_impl.cc:24] added target LDS to init manager Server
[2023-05-22 18:18:41.992][1042742][debug][init] [source/common/init/manager_impl.cc:49] init manager RTDS contains no targets
[2023-05-22 18:18:41.992][1042742][debug][init] [source/common/init/watcher_impl.cc:14] init manager RTDS initialized, notifying RTDS
[2023-05-22 18:18:41.992][1042742][info][runtime] [source/common/runtime/runtime_impl.cc:463] RTDS has finished initialization
[2023-05-22 18:18:41.992][1042742][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:244] continue initializing secondary clusters
[2023-05-22 18:18:41.992][1042742][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:169] maybe finish initialize state: 2
[2023-05-22 18:18:41.992][1042742][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:178] maybe finish initialize primary init clusters empty: true
[2023-05-22 18:18:41.992][1042742][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:193] maybe finish initialize secondary init clusters empty: true
[2023-05-22 18:18:41.992][1042742][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:220] maybe finish initialize cds api ready: true
[2023-05-22 18:18:41.992][1042742][info][upstream] [source/common/upstream/cluster_manager_impl.cc:222] cm init: initializing cds
[2023-05-22 18:18:41.992][1042742][debug][config] [source/common/config/grpc_mux_impl.cc:169] gRPC mux addWatch for type.googleapis.com/envoy.config.cluster.v3.Cluster
[2023-05-22 18:18:41.993][1042742][warning][main] [source/server/server.cc:794] there is no configured limit to the number of allowed active connections. Set a limit via the runtime key overload.global_downstream_max_connections
[2023-05-22 18:18:41.995][1042742][info][main] [source/server/server.cc:915] starting main dispatch loop
[2023-05-22 18:18:41.995][1042742][debug][connection] [source/common/network/connection_impl.cc:699] [C0] delayed connect error: 61
[2023-05-22 18:18:41.995][1042742][debug][connection] [source/common/network/connection_impl.cc:250] [C0] closing socket: 0
[2023-05-22 18:18:41.995][1042742][debug][client] [source/common/http/codec_client.cc:107] [C0] disconnect. resetting 0 pending requests
[2023-05-22 18:18:41.995][1042742][debug][pool] [source/common/conn_pool/conn_pool_base.cc:484] [C0] client disconnected, failure reason: delayed connect error: 61
[2023-05-22 18:18:41.995][1042742][debug][router] [source/common/router/router.cc:1208] [C0][S13139667687758040534] upstream reset: reset reason: connection failure, transport failure reason: delayed connect error: 61
[2023-05-22 18:18:41.996][1042742][debug][http] [source/common/http/async_client_impl.cc:105] async http request response headers (end_stream=true):
':status', '200'
'content-type', 'application/grpc'
'grpc-status', '14'
'grpc-message', 'upstream connect error or disconnect/reset before headers. reset reason: connection failure, transport failure reason: delayed connect error: 61'

[2023-05-22 18:18:41.996][1042742][debug][config] [./source/common/config/grpc_stream.h:170] StreamAggregatedResources gRPC config stream to xds_cluster closed: 14, upstream connect error or disconnect/reset before headers. reset reason: connection failure, transport failure reason: delayed connect error: 61
[2023-05-22 18:18:41.996][1042742][debug][config] [source/common/config/grpc_subscription_impl.cc:115] gRPC update for type.googleapis.com/envoy.config.cluster.v3.Cluster failed
[2023-05-22 18:18:41.996][1042742][debug][pool] [source/common/conn_pool/conn_pool_base.cc:454] invoking idle callbacks - is_draining_for_deletion_=false
```

The initial cluster named `xds_cluster` with URL `/envoy.service.discovery.v3.AggregatedDiscoveryService/StreamAggregatedResources` is used by the data plane to connect to the control plane. The HTTP header of `DiscoveryRequest` sent by the xDS client looks like this:

```http
':method', 'POST'
':path', '/envoy.service.discovery.v3.AggregatedDiscoveryService/StreamAggregatedResources'
':authority', 'xds_cluster'
':scheme', 'http'
'te', 'trailers'
'content-type', 'application/grpc'
'x-envoy-internal', 'true'
'x-forwarded-for', '192.168.1.248'
```

Then, start the xDS server, and we can observe:

```bash
[2023-05-22 19:51:31.123][1076786][debug][config] [source/common/config/grpc_mux_impl.cc:222] Received gRPC message for type.googleapis.com/envoy.config.cluster.v3.Cluster at version 1
[2023-05-22 19:51:31.123][1076786][debug][config] [source/common/config/grpc_mux_impl.cc:199] Pausing discovery requests for type.googleapis.com/envoy.config.cluster.v3.Cluster (previous count 0)
[2023-05-22 19:51:31.124][1076786][debug][config] [source/common/config/grpc_mux_impl.cc:199] Pausing discovery requests for type.googleapis.com/envoy.config.endpoint.v3.ClusterLoadAssignment (previous count 0)
[2023-05-22 19:51:31.124][1076786][debug][config] [source/common/config/grpc_mux_impl.cc:199] Pausing discovery requests for type.googleapis.com/envoy.config.endpoint.v3.LbEndpoint (previous count 0)
[2023-05-22 19:51:31.124][1076786][debug][config] [source/common/config/grpc_mux_impl.cc:199] Pausing discovery requests for type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret (previous count 0)
[2023-05-22 19:51:31.125][1076786][info][upstream] [source/common/upstream/cds_api_helper.cc:35] cds: add 1 cluster(s), remove 1 cluster(s)
[2023-05-22 19:51:31.126][1076786][debug][misc] [source/common/network/dns_resolver/dns_factory_util.cc:42] create Apple DNS resolver type: envoy.network.dns_resolver.apple in MacOS.
[2023-05-22 19:51:31.126][1076786][debug][misc] [source/common/network/dns_resolver/dns_factory_util.cc:81] create DNS resolver type: envoy.network.dns_resolver.apple
[2023-05-22 19:51:31.132][1076786][debug][config] [./source/common/http/filter_chain_helper.h:88]     upstream http filter #0
[2023-05-22 19:51:31.132][1076786][debug][config] [./source/common/http/filter_chain_helper.h:118]       name: envoy.filters.http.upstream_codec
[2023-05-22 19:51:31.132][1076786][debug][config] [./source/common/http/filter_chain_helper.h:121]     config: {"@type":"type.googleapis.com/envoy.extensions.filters.http.upstream_codec.v3.UpstreamCodec"}
[2023-05-22 19:51:31.134][1076786][debug][config] [source/common/config/grpc_mux_impl.cc:199] Pausing discovery requests for type.googleapis.com/envoy.config.cluster.v3.Cluster (previous count 1)
[2023-05-22 19:51:31.134][1076786][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:770] add/update cluster envoy_control starting warming
[2023-05-22 19:51:31.134][1076786][debug][upstream] [source/extensions/clusters/logical_dns/logical_dns_cluster.cc:108] starting async DNS resolution for www.google.com
[2023-05-22 19:51:31.134][1076786][debug][dns] [source/extensions/network/dns_resolver/apple/apple_dns_impl.cc:63] DNS resolution for www.google.com started
[2023-05-22 19:51:31.135][1076786][debug][upstream] [source/common/upstream/cds_api_helper.cc:52] cds: add/update cluster 'envoy_control'
[2023-05-22 19:51:31.135][1076786][info][upstream] [source/common/upstream/cds_api_helper.cc:72] cds: added/updated 1 cluster(s), skipped 0 unmodified cluster(s)
[2023-05-22 19:51:31.135][1076786][debug][config] [source/common/config/grpc_mux_impl.cc:206] Decreasing pause count on discovery requests for type.googleapis.com/envoy.config.endpoint.v3.ClusterLoadAssignment (previous count 1)
[2023-05-22 19:51:31.135][1076786][debug][config] [source/common/config/grpc_mux_impl.cc:206] Decreasing pause count on discovery requests for type.googleapis.com/envoy.config.endpoint.v3.LbEndpoint (previous count 1)
[2023-05-22 19:51:31.135][1076786][debug][config] [source/common/config/grpc_mux_impl.cc:206] Decreasing pause count on discovery requests for type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret (previous count 1)
[2023-05-22 19:51:31.136][1076786][debug][config] [source/common/config/grpc_subscription_impl.cc:85] gRPC config for type.googleapis.com/envoy.config.cluster.v3.Cluster accepted with 1 resources with version 1
[2023-05-22 19:51:31.136][1076786][debug][config] [source/common/config/grpc_mux_impl.cc:206] Decreasing pause count on discovery requests for type.googleapis.com/envoy.config.cluster.v3.Cluster (previous count 2)
[2023-05-22 19:51:31.136][1076786][debug][config] [source/common/config/grpc_mux_impl.cc:222] Received gRPC message for type.googleapis.com/envoy.config.listener.v3.Listener at version 1
[2023-05-22 19:51:31.136][1076786][debug][config] [source/common/config/grpc_mux_impl.cc:199] Pausing discovery requests for type.googleapis.com/envoy.config.listener.v3.Listener (previous count 0)
[2023-05-22 19:51:31.139][1076786][debug][config] [source/common/config/grpc_mux_impl.cc:199] Pausing discovery requests for type.googleapis.com/envoy.config.route.v3.RouteConfiguration (previous count 0)
[2023-05-22 19:51:31.139][1076786][debug][config] [source/common/config/grpc_mux_impl.cc:199] Pausing discovery requests for type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret (previous count 0)
[2023-05-22 19:51:31.142][1076786][debug][config] [source/extensions/listener_managers/listener_manager/listener_manager_impl.cc:468] begin add/update listener: name=listener_0 hash=1981995409266763902
[2023-05-22 19:51:31.142][1076786][debug][config] [source/extensions/listener_managers/listener_manager/listener_manager_impl.cc:505] use full listener update path for listener name=listener_0 hash=1981995409266763902
[2023-05-22 19:51:31.143][1076786][warning][config] [source/extensions/listener_managers/listener_manager/listener_impl.cc:1095] reuse_port was configured for TCP listener 'listener_0' and is being force disabled because Envoy is not running on Linux. See the documentation for more information.
[2023-05-22 19:51:31.153][1076786][debug][config] [source/extensions/listener_managers/listener_manager/listener_manager_impl.cc:88]   filter #0:
[2023-05-22 19:51:31.153][1076786][debug][config] [source/extensions/listener_managers/listener_manager/listener_manager_impl.cc:89]     name: envoy.filters.network.http_connection_manager
[2023-05-22 19:51:31.154][1076786][debug][config] [source/extensions/listener_managers/listener_manager/listener_manager_impl.cc:92]   config: {"@type":"type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager","stat_prefix":"ingress_http","http_filters":[{"name":"envoy.filters.http.router","typed_config":{"@type":"type.googleapis.com/envoy.extensions.filters.http.router.v3.Router"}}],"route_config":{"name":"local_route","virtual_hosts":[{"name":"local_service","domains":["*"],"routes":[{"match":{"prefix":"/"},"route":{"prefix_rewrite":"/robots.txt","cluster":"envoy_control","host_rewrite_literal":"www.google.com"}}]}]}}
[2023-05-22 19:51:31.172][1076786][debug][config] [./source/common/http/filter_chain_helper.h:88]     http filter #0
[2023-05-22 19:51:31.173][1076786][debug][config] [./source/common/http/filter_chain_helper.h:118]       name: envoy.filters.http.router
[2023-05-22 19:51:31.173][1076786][debug][config] [./source/common/http/filter_chain_helper.h:121]     config: {"@type":"type.googleapis.com/envoy.extensions.filters.http.router.v3.Router"}
[2023-05-22 19:51:31.176][1076786][debug][config] [source/extensions/listener_managers/listener_manager/filter_chain_manager_impl.cc:322] new fc_contexts has 1 filter chains, including 1 newly built
[2023-05-22 19:51:31.176][1076786][debug][config] [source/extensions/listener_managers/listener_manager/listener_impl.cc:156] Create listen socket for listener listener_0 on address 127.0.0.1:10000
[2023-05-22 19:51:31.176][1076786][debug][config] [source/extensions/listener_managers/listener_manager/listener_impl.cc:166] listener_0: Setting socket options succeeded
[2023-05-22 19:51:31.176][1076786][debug][config] [source/extensions/listener_managers/listener_manager/listener_impl.cc:109] Set listener listener_0 socket factory local address to 127.0.0.1:10000
[2023-05-22 19:51:31.176][1076786][debug][config] [source/extensions/listener_managers/listener_manager/listener_impl.cc:968] add warming listener: name=listener_0, hash=1981995409266763902, tag=1, address=127.0.0.1:10000
[2023-05-22 19:51:31.176][1076786][debug][misc] [source/extensions/listener_managers/listener_manager/listener_impl.cc:977] Initialize listener listener_0 local-init-manager.
[2023-05-22 19:51:31.176][1076786][debug][init] [source/common/init/manager_impl.cc:49] init manager Listener-local-init-manager listener_0 1981995409266763902 contains no targets
[2023-05-22 19:51:31.176][1076786][debug][init] [source/common/init/watcher_impl.cc:14] init manager Listener-local-init-manager listener_0 1981995409266763902 initialized, notifying Listener-local-init-watcher listener_0
[2023-05-22 19:51:31.176][1076786][debug][config] [source/extensions/listener_managers/listener_manager/listener_impl.cc:968] warm complete. updating active listener: name=listener_0, hash=1981995409266763902, tag=1, address=127.0.0.1:10000
[2023-05-22 19:51:31.176][1076786][info][upstream] [source/extensions/listener_managers/listener_manager/lds_api.cc:82] lds: add/update listener 'listener_0'
[2023-05-22 19:51:31.176][1076786][debug][config] [source/common/config/grpc_mux_impl.cc:206] Decreasing pause count on discovery requests for type.googleapis.com/envoy.config.route.v3.RouteConfiguration (previous count 1)
[2023-05-22 19:51:31.176][1076786][debug][config] [source/common/config/grpc_mux_impl.cc:206] Decreasing pause count on discovery requests for type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret (previous count 1)
[2023-05-22 19:51:31.176][1076786][debug][config] [source/common/config/grpc_subscription_impl.cc:85] gRPC config for type.googleapis.com/envoy.config.listener.v3.Listener accepted with 1 resources with version 1
[2023-05-22 19:51:31.176][1076786][debug][config] [source/common/config/grpc_mux_impl.cc:206] Decreasing pause count on discovery requests for type.googleapis.com/envoy.config.listener.v3.Listener (previous count 1)
[2023-05-22 19:51:31.176][1076786][debug][config] [source/common/config/grpc_mux_impl.cc:210] Resuming discovery requests for type.googleapis.com/envoy.config.listener.v3.Listener
[2023-05-22 19:51:31.176][1076786][debug][dns] [source/extensions/network/dns_resolver/apple/apple_dns_impl.cc:181] DNS resolver file event (1)
[2023-05-22 19:51:31.177][1076786][debug][dns] [source/extensions/network/dns_resolver/apple/apple_dns_impl.cc:311] DNS for www.google.com resolved with: flags=2[MoreComing=no, Add=yes], interface_index=0, error_code=0, hostname=www.google.com.
[2023-05-22 19:51:31.177][1076786][debug][dns] [source/extensions/network/dns_resolver/apple/apple_dns_impl.cc:334] Address to add address=142.250.190.36, ttl=243
[2023-05-22 19:51:31.177][1076786][debug][dns] [source/extensions/network/dns_resolver/apple/apple_dns_impl.cc:344] DNS Resolver flushing queries pending callback
[2023-05-22 19:51:31.177][1076786][debug][dns] [source/extensions/network/dns_resolver/apple/apple_dns_impl.cc:231] dns resolution for www.google.com completed with status 0
[2023-05-22 19:51:31.177][1076786][debug][upstream] [source/extensions/clusters/logical_dns/logical_dns_cluster.cc:116] async DNS resolution complete for www.google.com
[2023-05-22 19:51:31.177][1076786][debug][upstream] [source/common/upstream/upstream_impl.cc:451] transport socket match, socket default selected for host with address 142.250.190.36:443
[2023-05-22 19:51:31.178][1076786][debug][upstream] [source/extensions/clusters/logical_dns/logical_dns_cluster.cc:167] DNS refresh rate reset for www.google.com, refresh rate 5000 ms
[2023-05-22 19:51:31.178][1076786][debug][upstream] [source/common/upstream/upstream_impl.cc:1446] initializing Primary cluster envoy_control completed
[2023-05-22 19:51:31.178][1076786][debug][init] [source/common/init/manager_impl.cc:49] init manager Cluster envoy_control contains no targets
[2023-05-22 19:51:31.178][1076786][debug][init] [source/common/init/watcher_impl.cc:14] init manager Cluster envoy_control initialized, notifying ClusterImplBase
[2023-05-22 19:51:31.178][1076786][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:772] warming cluster envoy_control complete
[2023-05-22 19:51:31.178][1076786][debug][config] [source/common/config/grpc_mux_impl.cc:206] Decreasing pause count on discovery requests for type.googleapis.com/envoy.config.cluster.v3.Cluster (previous count 1)
[2023-05-22 19:51:31.178][1076786][debug][config] [source/common/config/grpc_mux_impl.cc:210] Resuming discovery requests for type.googleapis.com/envoy.config.cluster.v3.Cluster
[2023-05-22 19:51:31.178][1077078][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:1117] adding TLS cluster envoy_control
[2023-05-22 19:51:31.178][1077081][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:1117] adding TLS cluster envoy_control
[2023-05-22 19:51:31.178][1077077][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:1117] adding TLS cluster envoy_control
[2023-05-22 19:51:31.178][1077080][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:1117] adding TLS cluster envoy_control
[2023-05-22 19:51:31.178][1077081][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:1195] membership update for TLS cluster envoy_control added 1 removed 0
[2023-05-22 19:51:31.178][1077077][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:1195] membership update for TLS cluster envoy_control added 1 removed 0
[2023-05-22 19:51:31.178][1076786][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:1117] adding TLS cluster envoy_control
[2023-05-22 19:51:31.178][1077078][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:1195] membership update for TLS cluster envoy_control added 1 removed 0
[2023-05-22 19:51:31.178][1077080][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:1195] membership update for TLS cluster envoy_control added 1 removed 0
[2023-05-22 19:51:31.178][1076786][debug][upstream] [source/common/upstream/cluster_manager_impl.cc:1195] membership update for TLS cluster envoy_control added 1 removed 0
[2023-05-22 19:51:31.178][1076786][debug][dns] [source/extensions/network/dns_resolver/apple/apple_dns_impl.cc:235] Resolution for www.google.com completed (async)
[2023-05-22 19:51:31.178][1076786][debug][dns] [source/extensions/network/dns_resolver/apple/apple_dns_impl.cc:152] Destroying PendingResolution for www.google.com
[2023-05-22 19:51:31.178][1076786][debug][dns] [source/extensions/network/dns_resolver/apple/apple_dns_impl.cc:166] DNSServiceRefDeallocate individual sd ref
[2023-05-22 19:51:35.791][1076786][debug][main] [source/server/server.cc:265] flushing stats
[2023-05-22 19:51:36.177][1076786][debug][upstream] [source/extensions/clusters/logical_dns/logical_dns_cluster.cc:108] starting async DNS resolution for www.google.com
[2023-05-22 19:51:36.178][1076786][debug][dns] [source/extensions/network/dns_resolver/apple/apple_dns_impl.cc:63] DNS resolution for www.google.com started
[2023-05-22 19:51:36.179][1076786][debug][dns] [source/extensions/network/dns_resolver/apple/apple_dns_impl.cc:181] DNS resolver file event (1)
[2023-05-22 19:51:36.179][1076786][debug][dns] [source/extensions/network/dns_resolver/apple/apple_dns_impl.cc:311] DNS for www.google.com resolved with: flags=1073741826[MoreComing=no, Add=yes], interface_index=0, error_code=0, hostname=www.google.com.
[2023-05-22 19:51:36.179][1076786][debug][dns] [source/extensions/network/dns_resolver/apple/apple_dns_impl.cc:334] Address to add address=142.250.190.36, ttl=243
[2023-05-22 19:51:36.179][1076786][debug][dns] [source/extensions/network/dns_resolver/apple/apple_dns_impl.cc:344] DNS Resolver flushing queries pending callback
[2023-05-22 19:51:36.179][1076786][debug][dns] [source/extensions/network/dns_resolver/apple/apple_dns_impl.cc:231] dns resolution for www.google.com completed with status 0
[2023-05-22 19:51:36.179][1076786][debug][upstream] [source/extensions/clusters/logical_dns/logical_dns_cluster.cc:116] async DNS resolution complete for www.google.com
[2023-05-22 19:51:36.179][1076786][debug][upstream] [source/extensions/clusters/logical_dns/logical_dns_cluster.cc:167] DNS refresh rate reset for www.google.com, refresh rate 5000 ms
[2023-05-22 19:51:36.179][1076786][debug][dns] [source/extensions/network/dns_resolver/apple/apple_dns_impl.cc:235] Resolution for www.google.com completed (async)
[2023-05-22 19:51:36.179][1076786][debug][dns] [source/extensions/network/dns_resolver/apple/apple_dns_impl.cc:152] Destroying PendingResolution for www.google.com
[2023-05-22 19:51:36.179][1076786][debug][dns] [source/extensions/network/dns_resolver/apple/apple_dns_impl.cc:166] DNSServiceRefDeallocate individual sd ref
[2023-05-22 19:51:40.792][1076786][debug][main] [source/server/server.cc:265] flushing stats
```

As the `DiscoveryResponse` sent by the xDS server is being processed, new `DiscoveryRequest`s will be paused. Finally, access Envoy, and we can observe:

```bash
[2023-05-22 19:52:02.164][1077077][debug][conn_handler] [source/extensions/listener_managers/listener_manager/active_tcp_listener.cc:147] [C7] new connection from 127.0.0.1:50585
[2023-05-22 19:52:02.380][1077077][debug][http] [source/common/http/conn_manager_impl.cc:306] [C7] new stream
[2023-05-22 19:52:02.381][1077077][debug][http] [source/common/http/conn_manager_impl.cc:972] [C7][S3202891732645497323] request headers complete (end_stream=true):
':authority', 'http.domain.com'
':path', '/'
':method', 'GET'
'user-agent', 'curl/7.87.0'
'accept', '*/*'

[2023-05-22 19:52:02.381][1077077][debug][http] [source/common/http/conn_manager_impl.cc:955] [C7][S3202891732645497323] request end stream
[2023-05-22 19:52:02.383][1077077][debug][connection] [./source/common/network/connection_impl.h:92] [C7] current connecting state: false
[2023-05-22 19:52:02.384][1077077][debug][router] [source/common/router/router.cc:470] [C7][S3202891732645497323] cluster 'envoy_control' match for URL '/'
[2023-05-22 19:52:02.385][1077077][debug][router] [source/common/router/router.cc:678] [C7][S3202891732645497323] router decoding headers:
':authority', 'www.google.com'
':path', '/robots.txt'
':method', 'GET'
':scheme', 'https'
'user-agent', 'curl/7.87.0'
'accept', '*/*'
'x-forwarded-proto', 'https'
'x-request-id', 'e11df190-d825-451c-8bfb-586e5bcd5139'
'x-envoy-expected-rq-timeout-ms', '15000'
'x-envoy-original-path', '/'

[2023-05-22 19:52:02.385][1077077][debug][pool] [source/common/http/conn_pool_base.cc:78] queueing stream due to no available connections (ready=0 busy=0 connecting=0)
[2023-05-22 19:52:02.385][1077077][debug][pool] [source/common/conn_pool/conn_pool_base.cc:291] trying to create new connection
[2023-05-22 19:52:02.385][1077077][debug][pool] [source/common/conn_pool/conn_pool_base.cc:145] creating a new connection (connecting=0)
[2023-05-22 19:52:02.386][1077077][debug][connection] [./source/common/network/connection_impl.h:92] [C8] current connecting state: true
[2023-05-22 19:52:02.386][1077077][debug][client] [source/common/http/codec_client.cc:57] [C8] connecting
[2023-05-22 19:52:02.386][1077077][debug][connection] [source/common/network/connection_impl.cc:939] [C8] connecting to 142.250.190.36:443
[2023-05-22 19:52:02.387][1077077][debug][connection] [source/common/network/connection_impl.cc:958] [C8] connection in progress
[2023-05-22 19:52:02.407][1077077][debug][connection] [source/common/network/connection_impl.cc:688] [C8] connected
[2023-05-22 19:52:02.485][1077077][debug][client] [source/common/http/codec_client.cc:88] [C8] connected
[2023-05-22 19:52:02.485][1077077][debug][pool] [source/common/conn_pool/conn_pool_base.cc:328] [C8] attaching to next stream
[2023-05-22 19:52:02.485][1077077][debug][pool] [source/common/conn_pool/conn_pool_base.cc:182] [C8] creating stream
[2023-05-22 19:52:02.485][1077077][debug][router] [source/common/router/upstream_request.cc:581] [C7][S3202891732645497323] pool ready
[2023-05-22 19:52:02.485][1077077][debug][client] [source/common/http/codec_client.cc:139] [C8] encode complete
[2023-05-22 19:52:02.517][1077077][debug][router] [source/common/router/router.cc:1363] [C7][S3202891732645497323] upstream headers complete: end_stream=false
[2023-05-22 19:52:02.519][1077077][debug][http] [source/common/http/conn_manager_impl.cc:1588] [C7][S3202891732645497323] encoding headers via codec (end_stream=false):
':status', '200'
'accept-ranges', 'bytes'
'vary', 'Accept-Encoding'
'content-type', 'text/plain'
'cross-origin-resource-policy', 'cross-origin'
'cross-origin-opener-policy-report-only', 'same-origin; report-to="static-on-bigtable"'
'report-to', '{"group":"static-on-bigtable","max_age":2592000,"endpoints":[{"url":"https://csp.withgoogle.com/csp/report-to/static-on-bigtable"}]}'
'content-length', '7454'
'date', 'Tue, 23 May 2023 00:52:02 GMT'
'expires', 'Tue, 23 May 2023 00:52:02 GMT'
'cache-control', 'private, max-age=0'
'last-modified', 'Mon, 22 May 2023 21:30:00 GMT'
'x-content-type-options', 'nosniff'
'server', 'envoy'
'x-xss-protection', '0'
'alt-svc', 'h3=":443"; ma=2592000,h3-29=":443"; ma=2592000'
'x-envoy-upstream-service-time', '130'

[2023-05-22 19:52:02.519][1077077][debug][client] [source/common/http/codec_client.cc:126] [C8] response complete
[2023-05-22 19:52:02.519][1077077][debug][pool] [source/common/http/http1/conn_pool.cc:53] [C8] response complete
[2023-05-22 19:52:02.520][1077077][debug][pool] [source/common/conn_pool/conn_pool_base.cc:215] [C8] destroying stream: 0 remaining
[2023-05-22 19:52:02.522][1077077][debug][connection] [source/common/network/connection_impl.cc:656] [C7] remote close
[2023-05-22 19:52:02.522][1077077][debug][connection] [source/common/network/connection_impl.cc:250] [C7] closing socket: 0
[2023-05-22 19:52:02.522][1077077][debug][connection] [source/extensions/transport_sockets/tls/ssl_socket.cc:320] [C7] SSL shutdown: rc=1
[2023-05-22 19:52:02.522][1077077][debug][conn_handler] [source/extensions/listener_managers/listener_manager/active_stream_listener_base.cc:120] [C7] adding to cleanup list
[2023-05-22 19:52:05.797][1076786][debug][main] [source/server/server.cc:265] flushing stats
```

## Useful Links

[https://go.dev/blog/using-go-modules](https://go.dev/blog/using-go-modules)

[https://www.envoyproxy.io/docs/envoy/latest/intro/life_of_a_request](https://www.envoyproxy.io/docs/envoy/latest/intro/life_of_a_request)

[https://www.funnel-labs.io/2022/10/19/envoyproxy-5-securing-connections-with-https/](https://www.funnel-labs.io/2022/10/19/envoyproxy-5-securing-connections-with-https/)