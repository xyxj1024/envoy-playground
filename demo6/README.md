## Demo Setup

> Christian Posta: "The point is, you can build a workflow and process that statically configures the parts you need while use dynamic xDS services to discover the pieces you need at runtime. One of the reasons why you see different control-plane implementation is not everyone has a fully dynamic and fungible environment where all of the pieces should be dynamic. Adopt the level of dynamism that's most appropriate for your system given the existing constraints and available workflows."

### Development Environment

```bash
$ set | grep "MACHTYPE"
MACHTYPE=x86_64-apple-darwin21

$ go version
go version go1.20.5 darwin/amd64

$ envoy --version | grep -o '[0-9].[0-9]\+.[0-9]' | tail -n1
1.26.2
```

### How It Works

The `ingress` network is a special overlay network that facilitates load balancing among a service's node. When any swarm node receives a request on a published port, it hands that request off to the `IPVS` module, which keeps track of all the IP addresses participating in that service, selects one of them, and routes the request to it, over the `ingress` network.

The `docker_gwbridge` is a bridge network that connects the overlay networks (including the `ingress` network) to an individual Docker daemon's physical network. By default, each container a service is running is connected to its local Docker daemon host's `docker_gwbridge` network.

## Run Code

```bash
# Create overlay network and Envoy instances
.$(pwd)/deploy/scripts/deploy.sh

# Run control plane
go run envoy-swarm-control --debug

# Update Envoy services
docker service update \
    --label-add envoy.status.node-id=local_node_1 \
    --label-add envoy.listener.port=10001 \
    --label-add envoy.endpoint.port=80 \
    --label-add envoy.route.domain=example.com \
    --label-add envoy.route.upstream-host=www.google.com \
    envoy-1

docker service update \
    --label-add envoy.status.node-id=local_node_2 \
    --label-add envoy.listener.port=10002 \
    --label-add envoy.endpoint.port=80 \
    --label-add envoy.route.domain=example.com \
    --label-add envoy.route.upstream-host=www.wustl.edu \
    envoy-2

# Check Envoy logs
docker service logs -f envoy-1

docker service logs -f envoy-2
```